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
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// HostResolver resolves hostname to MidoNet Host ID.
type HostResolver struct {
	client *Client
}

// NewHostResolver creates a HostResolver.
func NewHostResolver(client *Client) *HostResolver {
	return &HostResolver{client}
}

// ResolveHost resolves a hostname to the corresponding MidoNet Host ID.
func (h *HostResolver) ResolveHost(hostname string) (*uuid.UUID, error) {
	clog := log.WithField("hostname", hostname)
	clog.Debug("Start resolving")
	hosts, err := listHosts(h.client)
	if err != nil {
		clog.WithError(err).Error("listHosts")
		return nil, err
	}
	clog.WithField("hosts", hosts).Debug("Got hosts")
	for _, h := range hosts {
		if h.Name == hostname {
			clog.WithField("ID", h.ID).Info("Resolved")
			return h.ID, nil
		}
	}
	clog.Info("No host found")
	return nil, fmt.Errorf("Host %s not found", hostname)
}

func listHosts(c *Client) ([]Host, error) {
	var hosts []Host
	_, err := c.List(&hosts)
	if err != nil {
		return nil, err
	}
	return hosts, nil
}
