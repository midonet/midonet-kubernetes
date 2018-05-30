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
	"net"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/google/uuid"
)

func ParseCIDR(s string) (*types.IPNet, error) {
	tmp, err := types.ParseCIDR(s)
	if err != nil {
		return nil, err
	}
	ip := types.IPNet(*tmp)
	return &ip, nil
}

func JumpRule(id *uuid.UUID, from *uuid.UUID, to *uuid.UUID) *Rule {
	return &Rule{
		Parent:      Parent{ID: from},
		ID:          id,
		Type:        "jump",
		JumpChainID: to,
	}
}

// https://docs.midonet.org/docs/v5.4/en/rest-api/content/resource-models.html

type HasParent interface {
	GetParent() *uuid.UUID
	SetParent(*uuid.UUID)
}

type Parent struct {
	ID *uuid.UUID `json:"-"`
}

func (p *Parent) GetParent() *uuid.UUID {
	return p.ID
}

func (p *Parent) SetParent(id *uuid.UUID) {
	p.ID = id
}

type PortRange struct {
	// Can't specify 0 explicitly but it should be ok for our usage
	Start int `json:"start,omitempty"`
	End   int `json:"end,omitempty"`
}

type NATTarget struct {
	// Can't specify 0 explicitly but it should be ok for our usage
	AddressFrom string `json:"addressFrom,omitempty"`
	AddressTo   string `json:"addressTo,omitempty"`
	PortFrom    int    `json:"portFrom,omitempty"`
	PortTo      int    `json:"portTo,omitempty"`
}

// https://docs.midonet.org/docs/v5.4/en/rest-api/content/bridge.html
type Bridge struct {
	midonetResource
	ID               *uuid.UUID `json:"id,omitempty"`
	TenantID         string     `json:"tenantId,omitempty"`
	Name             string     `json:"name,omitempty"`
	InboundFilterID  *uuid.UUID `json:"inboundFilterId,omitempty"`
	OutboundFilterID *uuid.UUID `json:"outboundFilterId,omitempty"`
}

func (_ *Bridge) MediaType() string {
	return "application/vnd.org.midonet.Bridge-v4+json"
}

func (res *Bridge) Path(op string) string {
	switch op {
	case "POST":
		return "/bridges"
	case "PUT", "DELETE", "GET":
		return fmt.Sprintf("/bridges/%s", res.ID)
	default:
		return ""
	}
}

// https://docs.midonet.org/docs/v5.4/en/rest-api/content/port.html
type Port struct {
	midonetResource
	Parent
	ID               *uuid.UUID     `json:"id,omitempty"`
	Type             string         `json:"type"`
	PortSubnet       []*types.IPNet `json:"portSubnet,omitempty"`
	PortMAC          HardwareAddr   `json:"portMac,omitempty"`
	InboundFilterID  *uuid.UUID     `json:"inboundFilterId,omitempty"`
	OutboundFilterID *uuid.UUID     `json:"outboundFilterId,omitempty"`
}

func (_ *Port) MediaType() string {
	return "application/vnd.org.midonet.Port-v3+json"
}

func (res *Port) Path(op string) string {
	switch op {
	case "POST":
		var parentType string
		switch res.Type {
		case "Bridge":
			parentType = "bridges"
		case "Router":
			parentType = "routers"
		}
		return fmt.Sprintf("/%s/%s/ports", parentType, res.Parent.ID)
	case "PUT", "DELETE", "GET":
		return fmt.Sprintf("/ports/%s", res.ID)
	default:
		return ""
	}
}

// https://docs.midonet.org/docs/v5.4/en/rest-api/content/port-link.html
type PortLink struct {
	midonetResource
	Parent
	PortID *uuid.UUID `json:"portId"`
	PeerID *uuid.UUID `json:"peerId"`
}

func (_ *PortLink) MediaType() string {
	return "application/vnd.org.midonet.PortLink-v1+json"
}

func (res *PortLink) Path(op string) string {
	switch op {
	case "POST", "DELETE":
		return fmt.Sprintf("/ports/%s/link", res.Parent.ID)
	default:
		return ""
	}
}

// https://docs.midonet.org/docs/v5.4/en/rest-api/content/route.html
type Route struct {
	midonetResource
	Parent
	ID               *uuid.UUID `json:"id,omitempty"`
	DstNetworkAddr   net.IP     `json:"dstNetworkAddr"`
	DstNetworkLength int        `json:"dstNetworkLength"`
	NextHopGateway   net.IP     `json:"nextHopGateway,omitempty"`
	NextHopPort      *uuid.UUID `json:"nextHopPort"`
	SrcNetworkAddr   net.IP     `json:"srcNetworkAddr"`
	SrcNetworkLength int        `json:"srcNetworkLength"`
	Type             string     `json:"type"`
}

func (_ *Route) MediaType() string {
	return "application/vnd.org.midonet.Route-v1+json"
}

func (res *Route) Path(op string) string {
	switch op {
	case "POST":
		return fmt.Sprintf("/routers/%s/routes", res.Parent.ID)
	case "DELETE", "GET":
		return fmt.Sprintf("/routes/%s", res.ID)
	default:
		return ""
	}
}

// https://docs.midonet.org/docs/v5.4/en/rest-api/content/chain.html
type Chain struct {
	midonetResource
	ID       *uuid.UUID `json:"id,omitempty"`
	TenantID string     `json:"tenantId,omitempty"`
	Name     string     `json:"name,omitempty"`
}

func (_ *Chain) MediaType() string {
	return "application/vnd.org.midonet.Chain-v1+json"
}

func (res *Chain) Path(op string) string {
	switch op {
	case "POST":
		return "/chains"
	case "DELETE", "GET":
		return fmt.Sprintf("/chains/%s", res.ID)
	default:
		return ""
	}
}

// https://docs.midonet.org/docs/v5.4/en/rest-api/content/rule.html
type Rule struct {
	midonetResource
	Parent
	ID           *uuid.UUID `json:"id,omitempty"`
	Type         string     `json:"type"`
	DLType       int        `json:"dlType,omitempty"`
	NWDstAddress string     `json:"nwDstAddress,omitempty"`
	NWDstLength  int        `json:"nwDstLength,omitempty"`
	NWProto      int        `json:"nwProto,omitempty"`
	NWSrcAddress string     `json:"nwSrcAddress,omitempty"`
	NWSrcLength  int        `json:"nwSrcLength,omitempty"`
	TPDst        *PortRange `json:"tpDst,omitempty"`
	TPSrc        *PortRange `json:"tpSrc,omitempty"`

	// JUMP
	JumpChainID *uuid.UUID `json:"jumpChainId,omitempty"`

	// DNAT, SNAT, REV_DNAT, REV_DNAT
	FlowAction string `json:"flowAction,omitempty"`

	// DNAT, SNAT
	NATTargets *[]NATTarget `json:"natTargets,omitempty"`
}

func (_ *Rule) MediaType() string {
	return "application/vnd.org.midonet.Rule-v2+json"
}

func (res *Rule) Path(op string) string {
	switch op {
	case "POST":
		return fmt.Sprintf("/chains/%s/rules", res.Parent.ID)
	case "DELETE", "GET":
		return fmt.Sprintf("/rules/%s", res.ID)
	default:
		return ""
	}
}

// https://docs.midonet.org/docs/v5.4/en/rest-api/content/host.html

type Host struct {
	midonetResource
	ID   *uuid.UUID `json:"id,omitempty"`
	Name string     `json:"name,omitempty"`
}

func (_ *Host) CollectionMediaType() string {
	return "application/vnd.org.midonet.collection.Host-v3+json"
}

func (_ *Host) Path(op string) string {
	switch op {
	case "LIST":
		return "/hosts"
	default:
		return ""
	}
}

// https://docs.midonet.org/docs/v5.4/en/rest-api/content/host-interface-port.html

type HostInterfacePort struct {
	midonetResource
	Parent
	HostID        *uuid.UUID `json:"hostId,omitempty"`
	PortID        *uuid.UUID `json:"portId,omitempty"`
	InterfaceName string     `json:"interfaceName"`
}

func (_ *HostInterfacePort) MediaType() string {
	return "application/vnd.org.midonet.HostInterfacePort-v1+json"
}

func (res *HostInterfacePort) Path(op string) string {
	switch op {
	case "POST":
		return fmt.Sprintf("/hosts/%s/ports", res.Parent.ID)
	case "DELETE", "GET":
		return fmt.Sprintf("/hosts/%s/ports/%s", res.Parent.ID, res.PortID)
	default:
		return ""
	}
}
