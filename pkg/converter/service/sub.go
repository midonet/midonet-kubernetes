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

package service

import (
	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/midonet"
)

type servicePort struct {
	portKey string
	ip      string
	proto   int
	port    int
}

func (s *servicePort) Convert(key string, config *midonet.Config) ([]converter.BackendResource, error) {
	svcsChainID := converter.ServicesChainID(config)
	jumpRuleID := converter.IDForKey("ServicePortSub", key)
	portChainID := converter.IDForKey("ServicePort", s.portKey)
	return []converter.BackendResource{
		&midonet.Rule{
			Parent:       midonet.Parent{ID: &svcsChainID},
			ID:           &jumpRuleID,
			DLType:       0x800,
			NWDstAddress: s.ip,
			NWDstLength:  32,
			NWProto:      s.proto,
			TPDst:        &midonet.PortRange{Start: s.port, End: s.port},
			Type:         "jump",
			JumpChainID:  &portChainID,
		},
	}, nil
}
