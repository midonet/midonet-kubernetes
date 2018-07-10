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

package pod

import (
	"net"

	"github.com/google/uuid"

	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/midonet"
)

type PortARP struct {
	BridgeID uuid.UUID
	IP       net.IP
	MAC      net.HardwareAddr
}

func (p *PortARP) Convert(key converter.Key, config *converter.Config) ([]converter.BackendResource, error) {
	return []converter.BackendResource{
		&midonet.IPv4MACPair{
			Parent: midonet.Parent{ID: &p.BridgeID},
			IP:     p.IP,
			MAC:    midonet.HardwareAddr(p.MAC),
		},
	}, nil
}
