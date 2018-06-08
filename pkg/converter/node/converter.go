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
	"fmt"
	"net"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"

	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/midonet"
)

func IDForKey(key string) uuid.UUID {
	return converter.IDForKey("Node", key)
}

func PortIDForKey(key string) uuid.UUID {
	baseID := IDForKey(key)
	return converter.SubID(baseID, "Node Port")
}

type nodeConverter struct{}

func newNodeConverter() converter.Converter {
	return &nodeConverter{}
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

func nodeAddresses(nodeKey string, routerPortID uuid.UUID, nodeIP net.IP, as []v1.NodeAddress) converter.SubResourceMap {
	subs := make(converter.SubResourceMap)
	for _, a := range as {
		typ := a.Type
		if typ != v1.NodeExternalIP && typ != v1.NodeInternalIP {
			continue
		}
		ip := net.ParseIP(a.Address)
		if ip == nil {
			// REVISIT: can this happen?
			log.WithFields(log.Fields{
				"node":    nodeKey,
				"address": a.Address,
			}).Fatal("Unparsable Node Address")
		}
		name := fmt.Sprintf("%s/%s/%s", nodeKey, typ, ip)
		subs[name] = &nodeAddress{
			routerPortID: routerPortID,
			nodeIP:       nodeIP,
			ip:           ip,
		}
	}
	return subs
}
