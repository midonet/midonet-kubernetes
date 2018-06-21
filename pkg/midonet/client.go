// Copyright (C) 2018 Midokura SARL.
// All rights reserved.
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package midonet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Client is a MidoNet API client.
type Client struct {
	config *Config
	token  string
}

// NewClient creates a Client.
func NewClient(config *Config) *Client {
	return &Client{
		config: config,
	}
}

func getZeroValue(res APIResource) APIResource {
	return reflect.New(reflect.TypeOf(res).Elem()).Interface().(APIResource)
}

func (c *Client) exists(origRes APIResource) (bool, error) {
	res := getZeroValue(origRes)
	resp, err := c.get(origRes, res)
	if err != nil {
		return false, err
	}
	log.WithFields(log.Fields{
		"resp": resp,
		"body": res,
	}).Debug("Get result")
	if resp.StatusCode/100 == 2 {
		// REVISIT: we can check the contents
		return true, nil
	}
	if resp.StatusCode == 404 {
		return false, nil
	}
	return false, fmt.Errorf("Unexpected status %d", resp.StatusCode)
}

// Push creates or updates the given resources on MidoNet API.
func (c *Client) Push(resources []APIResource) error {
	for _, res := range resources {
		// REVISIT: maybe we should save updates (and thus zk and
		// midolman loads) by performing GET and compare first.
		// Or we can make the MidoNet API detect and ignore no-op updates.
		resp, body, err := c.post(res)
		if err != nil {
			return err
		}
		if resp.StatusCode == 404 || resp.StatusCode == 400 {
			// There are a few cases we can see 404 here.
			// - The resource is HasParent and the parent has not been
			//   created yet
			// - The resource has a reference to the other resources (e.g.
			//   filter chains for a Bridge) and they have not been created
			//   yet
			// Also, MidoNet API returns 400 in a similar cases.
			// - When the port referenced by Route.nextHopPort doesn't exist.
			//   (ROUTE_NEXT_HOP_PORT_NOT_NULL)
			log.WithFields(log.Fields{
				"resource": res,
			}).Info("Referent doesn't exist yet?")
			return fmt.Errorf("Referent doesn't exist yet?")
		}
		if resp.StatusCode == 409 {
			if res.Path("PUT") != "" {
				resp, body, err = c.put(res)
				if err != nil {
					return err
				}
				if resp.StatusCode == 409 {
					if _, ok := res.(*TunnelZone); ok {
						// Workaound for UNIQUE_TUNNEL_ZONE_NAME_TYPE issue.
						// https://midonet.atlassian.net/browse/MNA-1293
						continue
					}
				}
			} else {
				if res.Path("GET") != "" {
					exists, err := c.exists(res)
					if err != nil {
						return err
					}
					if !exists {
						// assume a transient error
						return fmt.Errorf("Unexpected 409")
					}
				}
				// assume 409 meant ok
				// REVISIT: confirm that the existing resource is
				// same enough as what we want.
				continue
			}
		}
		if resp.StatusCode/100 != 2 {
			log.WithFields(log.Fields{
				"statusCode": resp.StatusCode,
				"body":       body,
			}).Fatal("Unexpected status code")
		}
	}
	return nil
}

// Delete deletes the given resources on MidoNet API.
func (c *Client) Delete(resources []APIResource) error {
	for _, res := range resources {
		resp, body, err := c.doRequest("DELETE", res.Path("DELETE"), nil, "")
		if err != nil {
			return err
		}
		// Ignore 404 assuming it's ok.  Even if we're the only one making
		// MidoNet topology modifications, it happens e.g. when a removal
		// of a Chain cascade-deleted Rules.
		if resp.StatusCode/100 != 2 && resp.StatusCode != 404 {
			log.WithFields(log.Fields{
				"statusCode": resp.StatusCode,
				"body":       body,
			}).Fatal("Unexpected status code")
		}
	}
	return nil
}

func (c *Client) post(res APIResource) (*http.Response, string, error) {
	return c.doRequest("POST", res.Path("POST"), res, "")
}

func (c *Client) put(res APIResource) (*http.Response, string, error) {
	return c.doRequest("PUT", res.Path("PUT"), res, "")
}

func (c *Client) get(id, result APIResource) (*http.Response, error) {
	resp, body, err := c.doRequest("GET", id.Path("GET"), nil, id.MediaType())
	if err != nil {
		return resp, err
	}
	dec := json.NewDecoder(strings.NewReader(body))
	err = dec.Decode(result)
	return resp, err
}

// List gets the list of the given resources on MidoNet API.
// Note: the argument rs should be a pointer to an empty slice of
// the struct.  E.g. a pointer to []Host
func (c *Client) List(rs interface{}) (*http.Response, error) {
	// assumption: rs is a pointer to an array of ListableResource
	// E.g. *[]Host
	t := reflect.TypeOf(rs)
	et := t.Elem().Elem()
	p := reflect.New(et)
	r := p.Interface().(ListableResource)
	resp, body, err := c.doRequest("GET", r.Path("LIST"), nil, r.CollectionMediaType())
	if err != nil {
		return resp, err
	}
	dec := json.NewDecoder(strings.NewReader(body))
	err = dec.Decode(rs)
	return resp, err
}

func (c *Client) doRequest(method string, path string, res APIResource, respType string) (*http.Response, string, error) {
	resp, body, err := c.request(method, path, res, respType)
	if err != nil {
		return resp, body, err
	}
	if resp.StatusCode == 401 {
		err = c.login()
		if err != nil {
			return nil, "", err
		}
		resp, body, err = c.request(method, path, res, respType)
	}
	return resp, body, err
}

func (c *Client) request(method string, path string, res APIResource, respType string) (*http.Response, string, error) {
	req, err := c.prepareRequest(method, path, res, respType)
	if err != nil {
		return nil, "", err
	}
	if c.token != "" {
		req.Header.Add("X-Auth-Token", c.token)
	}
	return c.executeRequest(req)
}

func (c *Client) prepareRequest(method string, path string, res APIResource, respType string) (*http.Request, error) {
	url := c.config.api + path
	clog := log.WithFields(log.Fields{
		"method": method,
		"url":    url,
	})
	var body io.Reader
	if res != nil {
		data, err := json.Marshal(res)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
		clog = clog.WithField("requestBody", string(data))
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if res != nil {
		req.Header.Add("Content-Type", res.MediaType())
	}
	if respType != "" {
		req.Header.Add("Accept", respType)
	}
	clog.Debug("prepareRequest")
	return req, err
}

func (c *Client) executeRequest(req *http.Request) (*http.Response, string, error) {
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)
	clog := log.WithFields(log.Fields{
		"statusCode":   resp.StatusCode,
		"responseBody": string(respBody),
	})
	clog.Debug("executeRequest")
	return resp, string(respBody), nil
}

//  https://docs.midonet.org/docs/latest-en/rest-api/content/authentication-authorization.html

type tokenInfo struct {
	Key     string `json:"key"`
	Expires string `json:"expires"`
}

func (c *Client) login() error {
	user := c.config.username
	pass := c.config.password
	project := c.config.project
	req, err := c.prepareRequest("POST", "/login", nil, "")
	if err != nil {
		return err
	}
	req.SetBasicAuth(user, pass)
	req.Header.Add("X-Auth-Project", project)
	resp, body, err := c.executeRequest(req)
	if err != nil {
		return err
	}
	if resp.StatusCode/100 != 2 {
		log.WithField("statusCode", resp.StatusCode).Fatal("Login failure")
	}
	dec := json.NewDecoder(strings.NewReader(body))
	info := &tokenInfo{}
	err = dec.Decode(info)
	if err != nil {
		return err
	}
	log.WithField("tokenInfo", info).Info("login succeeded.")
	c.token = info.Key
	return err
}
