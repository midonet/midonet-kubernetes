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

package endpoints

import (
	"fmt"

	"k8s.io/api/core/v1"

	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/midonet"
)

type endpoint struct {
	endpointsKey string
	portName     string
	svcIP        string
	ip           string
	port         int
	protocol     v1.Protocol
}

func (ep *endpoint) portKey() string {
	// Note: portKey format should be consistent with the
	// service converter.
	return fmt.Sprintf("%s/%s", ep.endpointsKey, ep.portName)
}

func (ep *endpoint) Convert(epKey converter.Key, config *midonet.Config) ([]converter.BackendResource, error) {
	// REVISIT: An assumption here is that, if ServicePort.Name is empty,
	// the corresponding EndpointPort.Name is also empty.  It isn't clear
	// to me (yamamoto) from the documentation.
	portKey := ep.portKey()
	portChainID := converter.IDForKey("ServicePort", portKey)
	baseID := converter.IDForKey("Endpoint", epKey.Key())
	epChainID := baseID
	epJumpRuleID := converter.SubID(baseID, "Jump to Endpoint")
	epDNATRuleID := converter.SubID(baseID, "DNAT")
	epSNATRuleID := converter.SubID(baseID, "SNAT")
	return []converter.BackendResource{
		&midonet.Chain{
			ID:       &epChainID,
			Name:     fmt.Sprintf("KUBE-SEP-%s", epKey.Key()),
			TenantID: config.Tenant,
		},
		// REVISIT: kube-proxy implements load-balancing with its
		// equivalent of this rule, using iptables probabilistic
		// match.  We can probably implement something similar
		// here if the backend has the following functionalities.
		//
		//   1. probabilistic match
		//   2. applyIfExists equivalent
		//
		// For now, we just install a normal 100% matching rule.
		// It means that the endpoint which happens to have its
		// jump rule the earliest in the chain handles 100% of
		// traffic.
		&midonet.Rule{
			Parent:      midonet.Parent{ID: &portChainID},
			ID:          &epJumpRuleID,
			Type:        "jump",
			JumpChainID: &epChainID,
		},
		&midonet.Rule{
			Parent: midonet.Parent{ID: &epChainID},
			ID:     &epDNATRuleID,
			Type:   "dnat",
			NATTargets: &[]midonet.NATTarget{
				{
					AddressFrom: ep.ip,
					AddressTo:   ep.ip,
					PortFrom:    ep.port,
					PortTo:      ep.port,
				},
			},
			FlowAction: "accept",
		},
		// SNAT traffic from the endpoint itself. Otherwise,
		// the return traffic doesn't work.
		// Note: Endpoint IP might or might not belong to the cluster ip
		// range.  It can be external.
		//
		// The source IP to use for this purpose is somewhat arbitrary
		// and doesn't seem consistent among networking implementations.
		// We use the ClusterIP of the corresponding Service.
		// For example, kube-proxy uses iptables MASQUERADE target for
		// this purpose.  It means that the source IP of the outgoing
		// interface is chosen after an L3 routing decision.  With flannel,
		// it would be the address of the cni0 interface on the node.
		&midonet.Rule{
			Parent:       midonet.Parent{ID: &epChainID},
			ID:           &epSNATRuleID,
			Type:         "snat",
			DLType:       0x800,
			NWSrcAddress: ep.ip,
			NWSrcLength:  32,
			NATTargets: &[]midonet.NATTarget{
				{
					AddressFrom: ep.svcIP,
					AddressTo:   ep.svcIP,
					// REVISIT: arbitrary port range
					PortFrom: 30000,
					PortTo:   60000,
				},
			},
			FlowAction: "continue",
		},
	}, nil
}
