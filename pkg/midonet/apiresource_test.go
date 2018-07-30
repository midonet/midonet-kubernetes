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
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/midonet/midonet-kubernetes/pkg/apis/midonet/v1"
)

func TestToAPIWithParent(t *testing.T) {
	id, _ := uuid.Parse("e88a600b-93aa-11e8-bcfa-0015170bebef")
	obj := TunnelZoneHost{
		Parent: Parent{ID: &id},
	}
	actual, err := obj.ToAPI(&obj)
	if err != nil {
		t.Errorf("got error %v", err)
	}
	expected := &v1.BackendResource{
		Kind: "TunnelZoneHost",
		Parent: "e88a600b-93aa-11e8-bcfa-0015170bebef",
		Body: "{}",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestToAPIWithoutParent(t *testing.T) {
	obj := TunnelZone{}
	actual, err := obj.ToAPI(&obj)
	if err != nil {
		t.Errorf("got error %v", err)
	}
	expected := &v1.BackendResource{
		Kind: "TunnelZone",
		Parent: "",
		Body: "{}",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestFromAPIWithParent(t *testing.T) {
	r := &v1.BackendResource{
		Kind: "TunnelZoneHost",
		Parent: "e88a600b-93aa-11e8-bcfa-0015170bebef",
		Body: "{}",
	}
	actual, err := FromAPI(r)
	if err != nil {
		t.Errorf("got error %v", err)
	}
	id, _ := uuid.Parse("e88a600b-93aa-11e8-bcfa-0015170bebef")
	expected := &TunnelZoneHost{
		Parent: Parent{ID: &id},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestFromAPIWithoutParent(t *testing.T) {
	r := &v1.BackendResource{
		Kind: "TunnelZone",
		Parent: "",
		Body: "{}",
	}
	actual, err := FromAPI(r)
	if err != nil {
		t.Errorf("got error %v", err)
	}
	expected := &TunnelZone{}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}
