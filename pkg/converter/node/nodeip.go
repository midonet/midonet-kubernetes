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

	"github.com/containernetworking/plugins/pkg/ip"
)

type SubnetInfo struct {
	GatewayIP net.IPNet
	NodeIP    net.IPNet
	Subnet    net.IPNet
}

func GetSubnetInfo(podCIDR string) (*SubnetInfo, error) {
	// Use the first IP for the gateway.
	// Use the next one for the IP for the interface to connect the node.
	// This should be consistent with the Node converter.
	addr, subnet, err := net.ParseCIDR(podCIDR)
	if err != nil {
		return nil, err
	}
	gatewayIP := ip.NextIP(addr)
	nodeIP := ip.NextIP(gatewayIP)
	return &SubnetInfo{
		GatewayIP: net.IPNet{IP: gatewayIP, Mask: subnet.Mask},
		NodeIP:    net.IPNet{IP: nodeIP, Mask: subnet.Mask},
		Subnet:    *subnet,
	}, nil
}
