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

	"github.com/containernetworking/cni/pkg/types"
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

func nodeAddresses(nodeKey converter.Key, routerPortID uuid.UUID, nodeIP net.IP, as []v1.NodeAddress) converter.SubResourceMap {
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
		key := converter.Key{
			Kind: "Node-Address",
			Name: fmt.Sprintf("%s/%s/%s", nodeKey.Name, typ, ip),
		}
		subs[key] = &nodeAddress{
			routerPortID: routerPortID,
			nodeIP:       nodeIP,
			ip:           ip,
		}
	}
	return subs
}

func getTunnelZoneID(idString string, config *midonet.Config) (uuid.UUID, error) {
	if idString == "" {
		return converter.DefaultTunnelZoneID(config), nil
	}
	return uuid.Parse(idString)
}

func (c *nodeConverter) Convert(key converter.Key, obj interface{}, config *midonet.Config) ([]converter.BackendResource, converter.SubResourceMap, error) {
	baseID := IDForKey(key.Key())
	routerPortMAC := converter.MACForKey(key.Key())
	routerID := converter.ClusterRouterID(config)
	bridgeID := baseID
	bridgePortID := converter.SubID(baseID, "Bridge Port")
	nodePortID := PortIDForKey(key.Key())
	nodePortChainID := converter.SubID(baseID, "Node Port Chain")
	nodeSNATRuleID := converter.SubID(baseID, "Node Port SNAT Rule")
	routerPortID := converter.SubID(baseID, "Router Port")
	subnetRouteID := converter.SubID(baseID, "Route")
	spec := obj.(*v1.Node).Spec
	status := obj.(*v1.Node).Status
	meta := obj.(*v1.Node).ObjectMeta
	bridgeName := key.Key()
	si, err := GetSubnetInfo(spec.PodCIDR)
	if err != nil {
		log.WithField("node", obj).Fatal("Failed to parse PodCIDR")
	}
	routerPortSubnet := []*types.IPNet{
		{si.GatewayIP.IP, si.GatewayIP.Mask},
	}
	gatewayIP := si.GatewayIP.IP.String()
	nodeIP := si.NodeIP.IP
	subnetAddr := si.Subnet.IP
	subnetLen, _ := si.Subnet.Mask.Size()
	hostID, err := uuid.Parse(meta.Annotations[converter.HostIDAnnotation])
	if err != nil {
		// Drop the error as it isn't retriable.
		// (until the Node is updated again)
		return nil, nil, nil
	}
	mainChainID := converter.MainChainID(config)
	subs := nodeAddresses(key, routerPortID, nodeIP, status.Addresses)
	tunnelZoneID, err := getTunnelZoneID(meta.Annotations[converter.TunnelZoneIDAnnotation], config)
	if err == nil {
		tunnelEndpointIP := net.ParseIP(meta.Annotations[converter.TunnelEndpointIPAnnotation])
		if tunnelEndpointIP != nil {
			k := converter.Key{
				Kind: "Node-Tunnel-Endpoint",
				Name: fmt.Sprintf("%s/%s/%s/%s", key.Name, tunnelEndpointIP, tunnelZoneID, hostID),
			}
			subs[k] = &nodeTunnelEndpoint{
				tunnelZoneID:     tunnelZoneID,
				tunnelEndpointIP: tunnelEndpointIP,
				hostID:           hostID,
			}
		}
	}
	return []converter.BackendResource{
		&midonet.Bridge{
			ID:              &bridgeID,
			Name:            bridgeName,
			TenantID:        config.Tenant,
			InboundFilterID: &mainChainID,
		},
		&midonet.Port{
			Parent: midonet.Parent{ID: &bridgeID},
			ID:     &bridgePortID,
			Type:   "Bridge",
		},
		&midonet.Port{
			Parent:     midonet.Parent{ID: &routerID},
			ID:         &routerPortID,
			Type:       "Router",
			PortSubnet: routerPortSubnet,
			// If we leave portMac unspecified for POST, MidoNet API
			// automatically generates random portMac.
			// On the other hand, for PUT, it clears the portMac field.
			// I suspect the latter is a bug.  Use a deterministically
			// generated MAC address to avoid issues.
			// See https://midonet.atlassian.net/browse/MNA-1251
			PortMAC: midonet.HardwareAddr(routerPortMAC),
		},
		&midonet.Route{
			Parent:           midonet.Parent{ID: &routerID},
			ID:               &subnetRouteID,
			DstNetworkAddr:   subnetAddr,
			DstNetworkLength: subnetLen,
			SrcNetworkAddr:   net.ParseIP("0.0.0.0"),
			SrcNetworkLength: 0,
			NextHopPort:      &routerPortID,
			Type:             "Normal",
		},
		&midonet.PortLink{
			Parent: midonet.Parent{ID: &bridgePortID},
			// Do not specify portId to avoid a MidoNet bug.
			// See https://midonet.atlassian.net/browse/MNA-1249
			// PortID: &bridgePortID,
			PeerID: &routerPortID,
		},
		&midonet.Chain{
			ID:       &nodePortChainID,
			Name:     fmt.Sprintf("KUBE-NODE-%s", key.Key()),
			TenantID: config.Tenant,
		},
		&midonet.Port{
			Parent:           midonet.Parent{ID: &bridgeID},
			ID:               &nodePortID,
			Type:             "Bridge",
			OutboundFilterID: &nodePortChainID,
		},
		&midonet.HostInterfacePort{
			Parent:        midonet.Parent{ID: &hostID},
			HostID:        &hostID,
			PortID:        &nodePortID,
			InterfaceName: IFName(),
		},
		// In the out-filter chain of the Node port, SNAT traffic from
		// the Node IP.  (It's usually the traffic routed by the above
		// route for apiserver.)  Otherwise, the return traffic will not
		// come back to us and we can't REV_DNAT for the services.
		// REVISIT: Depending on the way we implement the external
		// connectivity, we may want to perform SNAT for other source IP
		// addresses.
		&midonet.Rule{
			Parent:       midonet.Parent{ID: &nodePortChainID},
			ID:           &nodeSNATRuleID,
			Type:         "snat",
			DLType:       0x800,
			NWSrcAddress: nodeIP.String(),
			NWSrcLength:  32,
			NATTargets: &[]midonet.NATTarget{
				{
					AddressFrom: gatewayIP,
					AddressTo:   gatewayIP,
					// REVISIT: arbitrary port range
					PortFrom: 30000,
					PortTo:   60000,
				},
			},
			FlowAction: "continue",
		},
	}, subs, nil
}
