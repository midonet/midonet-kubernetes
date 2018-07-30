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
)

func TestHardwareAddrMarshal(t *testing.T) {
	mac, _ := net.ParseMAC("01:23:45:67:89:ab")
	obj := HardwareAddr(mac)
	actual, err := json.Marshal(obj)
	if err != nil {
		t.Errorf("got error %v", err)
	}
	expected := "\"01:23:45:67:89:ab\""
	if string(actual) != expected {
		t.Errorf("got %v\nwant %v", string(actual), expected)
	}
}

func TestHardwareAddrUnmarshal(t *testing.T) {
	jsonblob := []byte("\"01:23:45:67:89:ab\"")
	obj := HardwareAddr{}
	err := json.Unmarshal(jsonblob, &obj)
	if err != nil {
		t.Errorf("got error %v", err)
	}
	mac, _ := net.ParseMAC("01:23:45:67:89:ab")
	expected := HardwareAddr(mac)
	if !reflect.DeepEqual(obj, expected) {
		t.Errorf("got %v\nwant %v", obj, expected)
	}
}
