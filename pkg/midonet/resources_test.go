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
	"encoding/json"
	"net"
	"reflect"
	"testing"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBridgeEmpty(t *testing.T) {
	obj := Bridge{}
	actual, err := json.Marshal(obj)
	if err != nil {
		t.Errorf("got error %v", err)
	}
	expected := "{}"
	if string(actual) != expected {
		t.Errorf("got %v\nwant %v", string(actual), expected)
	}
}

func TestBridgeUnknownField(t *testing.T) {
	// ignore unknown fields
	blob := []byte(`{ "foo": 1 }`)
	actual := Bridge{}
	err := json.Unmarshal(blob, &actual)
	if err != nil {
		t.Errorf("got error %v", err)
	}
	expected := Bridge{}
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestBridgeArray(t *testing.T) {
	blob := []byte(`[{ "name": "foo" }, { "name": "bar" }]`)
	var actual []Bridge
	err := json.Unmarshal(blob, &actual)
	if err != nil {
		t.Errorf("got error %v", err)
	}
	expected := []Bridge{
		{Name: "foo"},
		{Name: "bar"},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestRouterPort(t *testing.T) {
	blob := []byte(`{ "portSubnet": ["192.168.1.1/24", "10.0.1.9/8"] }`)
	var actual Port
	err := json.Unmarshal(blob, &actual)
	if err != nil {
		t.Errorf("got error %v", err)
	}
	ip, err := ParseCIDR("192.168.1.1/24")
	if err != nil {
		t.Errorf("got error %v", err)
	}
	ip2, err := ParseCIDR("10.0.1.9/8")
	if err != nil {
		t.Errorf("got error %v", err)
	}
	expected := Port{
		PortSubnet: []*types.IPNet{
			ip,
			ip2,
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestRoute(t *testing.T) {
	blob := []byte(`{ "dstNetworkAddr": "192.168.1.0", "dstNetworkLength": 24, "nextHopGateway": "10.1.1.2" }`)
	var actual Route
	err := json.Unmarshal(blob, &actual)
	if err != nil {
		t.Errorf("got error %v", err)
	}
	expected := Route{
		DstNetworkAddr:   net.ParseIP("192.168.1.0"),
		DstNetworkLength: 24,
		NextHopGateway:   net.ParseIP("10.1.1.2"),
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestHosts(t *testing.T) {
	blob := []byte(`[{"id":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"k","addresses":["fe80:0:0:0:0:11ff:fe00:1102","127.0.0.1","0:0:0:0:0:0:0:1","fe80:0:0:0:ecee:eeff:feee:eeee","10.0.0.9","fe80:0:0:0:f816:3eff:fec6:ef35","fe80:0:0:0:e4e8:89ff:feb3:3c6","fe80:0:0:0:0:11ff:fe00:1101","169.254.123.1","fe80:0:0:0:ecee:eeff:feee:eeee","fe80:0:0:0:870:f4ff:fee7:7f5c","fe80:0:0:0:5000:99ff:fedd:debe","fe80:0:0:0:78d3:fcff:fe8f:f1e","10.1.0.2","fe80:0:0:0:40dc:7cff:fe5b:2d81","fe80:0:0:0:d4a9:9cff:fe4f:41b7","172.17.0.1","fe80:0:0:0:42:8ff:fec2:8b4f","fe80:0:0:0:5416:a3ff:fee8:1f34","10.1.0.0","fe80:0:0:0:e475:8eff:fede:14d8"],"hostInterfaces":[{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"midorecirc-dp","mac":"02:00:11:00:11:02","mtu":65535,"type":"Physical","endpoint":"UNKNOWN","portType":null,"addresses":["fe80:0:0:0:0:11ff:fe00:1102"],"status":3,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/midorecirc-dp"},{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"lo","mac":"00:00:00:00:00:00","mtu":65536,"type":"Virtual","endpoint":"LOCALHOST","portType":null,"addresses":["127.0.0.1","0:0:0:0:0:0:0:1"],"status":3,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/lo"},{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"mido2dbe366aec7","mac":"ee:ee:ee:ee:ee:ee","mtu":65000,"type":"Physical","endpoint":"UNKNOWN","portType":null,"addresses":["fe80:0:0:0:ecee:eeff:feee:eeee"],"status":3,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/mido2dbe366aec7"},{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"gretap0","mac":"00:00:00:00:00:00","mtu":1462,"type":"Physical","endpoint":"UNKNOWN","portType":null,"addresses":[],"status":2,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/gretap0"},{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"ens3","mac":"fa:16:3e:c6:ef:35","mtu":1450,"type":"Physical","endpoint":"DATAPATH","portType":null,"addresses":["10.0.0.9","fe80:0:0:0:f816:3eff:fec6:ef35"],"status":3,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/ens3"},{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"tnvxlan-overlay","mac":"e6:e8:89:b3:03:c6","mtu":65485,"type":"Physical","endpoint":"UNKNOWN","portType":null,"addresses":["fe80:0:0:0:e4e8:89ff:feb3:3c6"],"status":3,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/tnvxlan-overlay"},{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"midorecirc-host","mac":"02:00:11:00:11:01","mtu":65535,"type":"Physical","endpoint":"UNKNOWN","portType":null,"addresses":["fe80:0:0:0:0:11ff:fe00:1101","169.254.123.1"],"status":3,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/midorecirc-host"},{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"midofacce085ea7","mac":"ee:ee:ee:ee:ee:ee","mtu":65000,"type":"Physical","endpoint":"UNKNOWN","portType":null,"addresses":["fe80:0:0:0:ecee:eeff:feee:eeee"],"status":3,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/midofacce085ea7"},{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"tnvxlan-recirc","mac":"0a:70:f4:e7:7f:5c","mtu":65485,"type":"Physical","endpoint":"UNKNOWN","portType":null,"addresses":["fe80:0:0:0:870:f4ff:fee7:7f5c"],"status":3,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/tnvxlan-recirc"},{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"tnvxlan-vtep","mac":"52:00:99:dd:de:be","mtu":65485,"type":"Physical","endpoint":"UNKNOWN","portType":null,"addresses":["fe80:0:0:0:5000:99ff:fedd:debe"],"status":3,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/tnvxlan-vtep"},{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"tnvxlan-fip64","mac":"7a:d3:fc:8f:0f:1e","mtu":65485,"type":"Physical","endpoint":"UNKNOWN","portType":null,"addresses":["fe80:0:0:0:78d3:fcff:fe8f:f1e"],"status":3,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/tnvxlan-fip64"},{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"v1","mac":"42:dc:7c:5b:2d:81","mtu":1500,"type":"Physical","endpoint":"UNKNOWN","portType":null,"addresses":["10.1.0.2","fe80:0:0:0:40dc:7cff:fe5b:2d81"],"status":3,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/v1"},{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"tngre-overlay","mac":"d6:a9:9c:4f:41:b7","mtu":65490,"type":"Physical","endpoint":"UNKNOWN","portType":null,"addresses":["fe80:0:0:0:d4a9:9cff:fe4f:41b7"],"status":3,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/tngre-overlay"},{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"docker0","mac":"02:42:08:c2:8b:4f","mtu":1500,"type":"Physical","endpoint":"UNKNOWN","portType":null,"addresses":["172.17.0.1","fe80:0:0:0:42:8ff:fec2:8b4f"],"status":3,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/docker0"},{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"midonet","mac":"e6:f6:20:2d:a2:ee","mtu":1500,"type":"Physical","endpoint":"UNKNOWN","portType":null,"addresses":[],"status":2,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/midonet"},{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"v0","mac":"56:16:a3:e8:1f:34","mtu":65000,"type":"Physical","endpoint":"UNKNOWN","portType":null,"addresses":["fe80:0:0:0:5416:a3ff:fee8:1f34"],"status":3,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/v0"},{"hostId":"dbb7065f-ab57-433c-92b6-84816a9e87be","name":"flannel.1","mac":"e6:75:8e:de:14:d8","mtu":1400,"type":"Physical","endpoint":"UNKNOWN","portType":null,"addresses":["10.1.0.0","fe80:0:0:0:e475:8eff:fede:14d8"],"status":3,"uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces/flannel.1"}],"alive":true,"floodingProxyWeight":1,"containerWeight":1,"containerLimit":-1,"enforceContainerLimit":false,"ports":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/ports","uri":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be","interfaces":"http://localhost:8181/midonet-api/hosts/dbb7065f-ab57-433c-92b6-84816a9e87be/interfaces"}]`)

	var actual []Host
	err := json.Unmarshal(blob, &actual)
	if err != nil {
		t.Errorf("got error %v", err)
	}
	id, err := uuid.Parse("dbb7065f-ab57-433c-92b6-84816a9e87be")
	if err != nil {
		t.Errorf("failed to parse uuid error %v", err)
	}
	expected := []Host{{
		ID:   &id,
		Name: "k",
	}}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestMACPortPair(t *testing.T) {
	mac, _ := net.ParseMAC("01:23:45:67:89:ab")
	portID, _ := uuid.Parse("dbb7065f-ab57-433c-92b6-84816a9e87be")
	res := &MACPort{
		MACAddr: HardwareAddr(mac),
		PortID:  &portID,
	}
	actual := res.macPortPair()
	expected := "01-23-45-67-89-ab_dbb7065f-ab57-433c-92b6-84816a9e87be"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestIP4MACPair(t *testing.T) {
	ip := net.ParseIP("192.168.1.1")
	mac, _ := net.ParseMAC("01:23:45:67:89:ab")
	res := &IPv4MACPair{
		IP:  ip,
		MAC: HardwareAddr(mac),
	}
	actual := res.ip4MACPair()
	expected := "192.168.1.1_01-23-45-67-89-ab"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestAPIResources(t *testing.T) {
	objs := []APIResource{
		&TunnelZone{},
		&TunnelZoneHost{},
		&Router{},
		&Bridge{},
		&Port{Type: "Router"},
		&Port{Type: "Bridge"},
		&PortLink{},
		&Route{},
		&Chain{},
		&Rule{},
		&HostInterfacePort{},
		&MACPort{},
		&IPv4MACPair{},
	}
	methods := []string{
		"GET",
		"POST",
		"PUT",
		"DELETE",
		"LIST",
	}
	for _, obj := range objs {
		typ := obj.MediaType()
		assert.Regexp(t, "^application/vnd\\.org\\.midonet\\..+\\+json$", typ)
		assert.NotRegexp(t, "^application/vnd\\.org\\.midonet\\.collection\\..+\\+json$", typ)
		allEmpty := true
		for _, method := range methods {
			path := obj.Path(method)
			assert.Regexp(t, "^(/[[:alpha:]_]+(/.+|)|)$", path)
			if path != "" {
				allEmpty = false
			}
		}
		assert.False(t, allEmpty)
		hasParent, ok := obj.(HasParent)
		if ok {
			pid, _ := uuid.Parse("2bdd2d6a-95ff-11e8-bcfa-0015170bebef")
			hasParent.SetParent(&pid)
			assert.Equal(t, hasParent.GetParent(), &pid)
			assert.Contains(t, obj.Path("POST"), pid.String())
		}
	}
}

func TestListableResources(t *testing.T) {
	objs := []ListableResource{
		&Host{},
	}
	for _, obj := range objs {
		typ := obj.CollectionMediaType()
		assert.Regexp(t, "^application/vnd\\.org\\.midonet\\.collection\\..+\\+json$", typ)
		path := obj.Path("LIST")
		assert.Regexp(t, "^/[[:alpha:]_]+$", path)
	}
}
