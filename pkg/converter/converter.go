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
	"github.com/midonet/midonet-kubernetes/pkg/apis/midonet/v1"
)

// BackendResource represents backend resources converted from k8s resources
type BackendResource interface {
	ToAPI(interface{}) (*v1.BackendResource, error)
}

func toAPI(r BackendResource) (*v1.BackendResource, error) {
	b, err := r.ToAPI(r)
	return b, err
}

// Converter converts a Kubernetes resource to zero or more backend resources.
type Converter interface {
	Convert(key Key, obj interface{}, config *Config) ([]BackendResource, SubResourceMap, error)
}
