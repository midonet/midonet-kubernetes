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

func (c *nodeConverter) Convert(key string, obj interface{}, config *midonet.Config) ([]converter.BackendResource, converter.SubResourceMap, error) {
	baseID := IDForKey(key)
	routerPortMAC := converter.MACForKey(key)
	routerID := config.ClusterRouter
	bridgeID := baseID
	bridgePortID := converter.SubID(baseID, "Bridge Port")
	nodePortID := PortIDForKey(key)
	routerPortID := converter.SubID(baseID, "Router Port")
	subnetRouteID := converter.SubID(baseID, "Route")
	apiRouteID := converter.SubID(baseID, "APIRoute")
	spec := obj.(*v1.Node).Spec
	meta := obj.(*v1.Node).ObjectMeta
	bridgeName := key
	si, err := GetSubnetInfo(spec.PodCIDR)
	if err != nil {
		log.WithField("node", obj).Fatal("Failed to parse PodCIDR")
	}
	routerPortSubnet := []*types.IPNet{
		{si.GatewayIP.IP, si.GatewayIP.Mask},
	}
	apiSubnetAddr := config.KubernetesAPISubnet.IP
	apiSubnetLen, _ := config.KubernetesAPISubnet.Mask.Size()
	subnetAddr := si.Subnet.IP
	subnetLen, _ := si.Subnet.Mask.Size()
	hostID, err := uuid.Parse(meta.Annotations[converter.HostIDAnnotation])
	if err != nil {
		// Drop the error as it isn't retriable.
		// (until the Node is updated again)
		return nil, nil, nil
	}
	mainChainID := converter.MainChainID(config)
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
		&midonet.Port{
			Parent: midonet.Parent{ID: &bridgeID},
			ID:     &nodePortID,
			Type:   "Bridge",
		},
		&midonet.HostInterfacePort{
			Parent:        midonet.Parent{ID: &hostID},
			HostID:        &hostID,
			PortID:        &nodePortID,
			InterfaceName: IFName(),
		},
		// Forward the apiserver traffic to the Node IP, assuming that
		// the node network can forward it to the apiserver.
		// REVISIT: Probably this should be optional as it might not be
		// appropriate for every deployments.
		&midonet.Route{
			Parent:           midonet.Parent{ID: &routerID},
			ID:               &apiRouteID,
			DstNetworkAddr:   apiSubnetAddr,
			DstNetworkLength: apiSubnetLen,
			SrcNetworkAddr:   subnetAddr,
			SrcNetworkLength: subnetLen,
			NextHopPort:      &routerPortID,
			NextHopGateway:   si.NodeIP.IP,
			Type:             "Normal",
		},
	}, nil, nil
}
