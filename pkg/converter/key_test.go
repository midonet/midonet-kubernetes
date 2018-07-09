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
	"fmt"
	"testing"
)

func TestNewKeyFromClientKey(t *testing.T) {
	actual, err := newKeyFromClientKey("Sdn", "mido/net")
	if err != nil {
		t.Errorf("unexpected error %v", err)
		t.FailNow()
	}
	expected := Key{
		Kind:      "Sdn",
		Namespace: "mido",
		Name:      "net",
	}
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestNewKeyFromClientKeyWithoutNamespace(t *testing.T) {
	actual, err := newKeyFromClientKey("Sdn", "midonet")
	if err != nil {
		t.Errorf("unexpected error %v", err)
		t.FailNow()
	}
	expected := Key{
		Kind:      "Sdn",
		Namespace: "",
		Name:      "midonet",
	}
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestNewKeyFromClientKeyWithBrokenKey(t *testing.T) {
	_, err := newKeyFromClientKey("Sdn", "mi/do/net")
	if err == nil {
		t.Errorf("unexpected success")
	}
}

func TestKey(t *testing.T) {
	k := Key{
		Kind:      "Hoge-Fuga",
		Namespace: "foo",
		Name:      "bar",
	}
	actual := k.Key()
	expected := "foo/bar"
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}

func TestTranslationName(t *testing.T) {
	k := Key{
		Kind:      "Hoge-Fuga",
		Namespace: "foo",
		Name:      "bar",
	}
	actual := k.translationName()
	expected := fmt.Sprintf("hoge-fuga.%s.bar", TranslationVersion)
	if actual != expected {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}
