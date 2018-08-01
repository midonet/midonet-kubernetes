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
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/midonet/midonet-kubernetes/pkg/converter"
)

var (
	nodeWithTunnelZone = &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				converter.HostIDAnnotation:           "44FB4381-C99D-4389-86C3-FDCA765BCBDE",
				converter.TunnelZoneIDAnnotation:     "8B80D6B7-F04B-4075-AB5D-0C4D5C85D15E",
				converter.TunnelEndpointIPAnnotation: "192.2.0.99",
			},
		},
		Spec: v1.NodeSpec{
			PodCIDR: "10.1.2.0/24",
		},
	}

	nodeWithMAC = &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				converter.HostIDAnnotation: "44FB4381-C99D-4389-86C3-FDCA765BCBDE",
				converter.MACAnnotation:    "99:88:77:66:55:44",
			},
		},
		Spec: v1.NodeSpec{
			PodCIDR: "10.1.2.0/24",
		},
	}

	nodeWithAddresses = &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				converter.HostIDAnnotation: "44FB4381-C99D-4389-86C3-FDCA765BCBDE",
			},
		},
		Spec: v1.NodeSpec{
			PodCIDR: "10.1.2.0/24",
		},
		Status: v1.NodeStatus{
			Addresses: []v1.NodeAddress{
				{Type: v1.NodeHostName, Address: "awesome"},
				{Type: v1.NodeExternalIP, Address: "192.2.0.9"},
				{Type: v1.NodeInternalIP, Address: "192.2.0.10"},
			},
		},
	}
)

func TestConverterWithTunnelZone(t *testing.T) {
	key := converter.Key{
		Kind:      "Node",
		Namespace: "foo",
		Name:      "awesome-node",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &nodeConverter{}
	rs, subs, err := c.Convert(key, nodeWithTunnelZone, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 9)
	assert.Len(t, subs, 1)
	assert.Contains(t, subs, converter.Key{
		Kind: "Node-Tunnel-Endpoint",
		Name: "awesome-node/192.2.0.99/8b80d6b7-f04b-4075-ab5d-0c4d5c85d15e/44fb4381-c99d-4389-86c3-fdca765bcbde",
	})
}

func TestConverterWithMAC(t *testing.T) {
	key := converter.Key{
		Kind:      "Node",
		Namespace: "foo",
		Name:      "awesome-node",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &nodeConverter{}
	rs, subs, err := c.Convert(key, nodeWithMAC, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 9)
	assert.Len(t, subs, 2)
	assert.Contains(t, subs, converter.Key{
		Kind:        "Node-ARP",
		Name:        "awesome-node/ip/10.1.2.2/998877665544",
		Unversioned: true, // XXX
	})
	assert.Contains(t, subs, converter.Key{
		Kind:        "Node-MAC",
		Name:        "awesome-node/mac/998877665544",
		Unversioned: true, // XXX
	})
}

func TestConverterWithAddresses(t *testing.T) {
	key := converter.Key{
		Kind:      "Node",
		Namespace: "foo",
		Name:      "awesome-node",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &nodeConverter{}
	rs, subs, err := c.Convert(key, nodeWithAddresses, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 9)
	assert.Len(t, subs, 2)
	assert.Contains(t, subs, converter.Key{
		Kind: "Node-Address",
		Name: "awesome-node/ExternalIP/192.2.0.9",
	})
	assert.Contains(t, subs, converter.Key{
		Kind: "Node-Address",
		Name: "awesome-node/InternalIP/192.2.0.10",
	})
}
