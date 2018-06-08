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

package node

import (
	"net"

	"github.com/google/uuid"

	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/midonet"
)

type nodeAddress struct {
	routerPortID uuid.UUID
	nodeIP       net.IP
	ip           net.IP
}

func (i *nodeAddress) Convert(key string, config *midonet.Config) ([]converter.BackendResource, error) {
	routerID := config.ClusterRouter
	routeID := converter.IDForKey("Node Address", key)
	return []converter.BackendResource{
		// Forward the traffic to Node.Status.Addresses to the Node IP,
		// assuming that the node network can forward it.
		&midonet.Route{
			Parent:           midonet.Parent{ID: &routerID},
			ID:               &routeID,
			DstNetworkAddr:   i.ip,
			DstNetworkLength: 32,
			SrcNetworkAddr:   net.ParseIP("0.0.0.0"),
			SrcNetworkLength: 0,
			NextHopPort:      &i.routerPortID,
			NextHopGateway:   i.nodeIP,
			Type:             "Normal",
		},
	}, nil
}
