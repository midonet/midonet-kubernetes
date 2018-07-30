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

package converter

import (
	"testing"
)

func TestMakeDNS(t *testing.T) {
	actual := makeDNS("foo.bar")
	expected := "foo.bar"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestMakeDNSWithCapital(t *testing.T) {
	actual := makeDNS("Foo.Bar")
	expected := "foo.bar-84625620bc9d2c2a323e78a21663b3b1efe9c35f"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestMakeDNSWithSlash(t *testing.T) {
	actual := makeDNS("foo/bar")
	expected := "foo-bar-17cdeaefa5cc6022481c824e15a47a7726f593dd"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}
