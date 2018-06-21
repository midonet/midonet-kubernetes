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
	"strings"

	reflect2 "github.com/modern-go/reflect2"
)

// The purpose of the following functions are to manage the struct types in
// resources.go.  They are not intended to be generic at all.
// These are merely hacks to avoid having explicit registrations.

// TypeNameForObject returns the type name of the interface.
// E.g. if midonet.Bridge{} is given, this function returns "Bridge".
func TypeNameForObject(obj interface{}) string {
	t := reflect.TypeOf(obj)
	fullname := t.String()
	sep := strings.Split(fullname, ".")
	return sep[len(sep)-1]
}

// ObjectByTypeName returns a zero value of the struct with the given name.
// E.g. if "Bridge" is given, this function returns midonet.Bridge{}.
// Note: This implementation is evil. If it turned out to be a problem,
// replace it with a map-based solution.
func ObjectByTypeName(name string) interface{} {
	t := reflect2.TypeByName("midonet." + name)
	return reflect.New(t.Type1()).Interface()
}
