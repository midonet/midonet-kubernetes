package midonet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	config *Config
}

func NewClient(config *Config) *Client {
	return &Client{config}
}

func (c *Client) Push(resources []APIResource) error {
	for _, res := range resources {
		// REVISIT: maybe we should save updates (and thus zk and
		// midolman loads) by performing GET and compare first.
		// Or we can make the MidoNet API detect and ignore no-op updates.
		resp, err := c.post(res)
		if err != nil {
			return err
		}
		if resp.StatusCode == 404 {
			if _, ok := res.(HasParent); ok {
				log.Info("Parent doesn't exist yet?")
				return fmt.Errorf("Parent doesn't exist yet?")
			}
		}
		if resp.StatusCode == 409 {
			if res.Path("PUT") != "" {
				resp, err = c.put(res)
				if err != nil {
					return err
				}
			} else {
				// assume 409 meant ok
				continue
			}
		}
		if resp.StatusCode/100 != 2 {
			log.WithField("statusCode", resp.StatusCode).Fatal("Unexpected status code")
		}
	}
	return nil
}

func (c *Client) Delete(resources []APIResource) error {
	for _, res := range resources {
		resp, err := c.doRequest("DELETE", res.Path("DELETE"), nil)
		if err != nil {
			return err
		}
		// Ignore 404 assuming it's ok.  Even if we're the only one making
		// MidoNet topology modifications, it happens e.g. when a removal
		// of a Chain cascade-deleted Rules.
		if resp.StatusCode/100 != 2 && resp.StatusCode != 404 {
			log.WithField("statusCode", resp.StatusCode).Fatal("Unexpected status code")
		}
	}
	return nil
}

func (c *Client) post(res APIResource) (*http.Response, error) {
	return c.doRequest("POST", res.Path("POST"), res)
}

func (c *Client) put(res APIResource) (*http.Response, error) {
	return c.doRequest("PUT", res.Path("PUT"), res)
}

func (c *Client) doRequest(method string, path string, res APIResource) (*http.Response, error) {
	url := c.config.API + path
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

	// TODO: login
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)
	clog = clog.WithFields(log.Fields{
		"statusCode":   resp.StatusCode,
		"responseBody": string(respBody),
	})
	clog.Info("Do")
	return resp, nil
}
