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

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"

	"github.com/midonet/midonet-kubernetes/pkg/converter"
)

func TestConverterUnsupportedType(t *testing.T) {
	key := converter.Key{
		Kind:      "Service",
		Namespace: "foo",
		Name:      "bar",
	}
	obj := &v1.Service{
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceTypeLoadBalancer,
			ClusterIP: "192.2.0.1",
			Ports: []v1.ServicePort{
				{
					Name:     "cat",
					Protocol: "UDP",
					Port:     8000,
				},
				{
					Name:     "dog",
					Protocol: "TCP",
					Port:     200,
				},
			},
		},
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	converter := &serviceConverter{}
	rs, subs, err := converter.Convert(key, obj, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 0)
	assert.Len(t, subs, 0)
}

func TestConverterEmptyClusterIP(t *testing.T) {
	key := converter.Key{
		Kind:      "Node",
		Namespace: "foo",
		Name:      "bar",
	}
	obj := &v1.Service{
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: "",
		},
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	converter := &serviceConverter{}
	rs, subs, err := converter.Convert(key, obj, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 0)
	assert.Len(t, subs, 0)
}

func TestConverterClusterIPNone(t *testing.T) {
	key := converter.Key{
		Kind:      "Node",
		Namespace: "foo",
		Name:      "bar",
	}
	obj := &v1.Service{
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: v1.ClusterIPNone,
			Ports: []v1.ServicePort{
				{
					Name:     "cat",
					Protocol: "UDP",
					Port:     8000,
				},
				{
					Name:     "dog",
					Protocol: "TCP",
					Port:     200,
				},
			},
		},
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	converter := &serviceConverter{}
	rs, subs, err := converter.Convert(key, obj, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 0)
	assert.Len(t, subs, 0)
}

func TestConverterNoPorts(t *testing.T) {
	key := converter.Key{
		Kind:      "Node",
		Namespace: "foo",
		Name:      "bar",
	}
	obj := &v1.Service{
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: "192.2.0.1",
		},
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	converter := &serviceConverter{}
	rs, subs, err := converter.Convert(key, obj, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 0)
	assert.Len(t, subs, 0)
}

func TestConverter(t *testing.T) {
	key := converter.Key{
		Kind:      "Node",
		Namespace: "foo",
		Name:      "bar",
	}
	obj := &v1.Service{
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: "192.2.0.1",
			Ports: []v1.ServicePort{
				{
					Name:     "cat",
					Protocol: "UDP",
					Port:     8000,
				},
				{
					Name:     "dog",
					Protocol: "TCP",
					Port:     200,
				},
			},
		},
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &serviceConverter{}
	rs, subs, err := c.Convert(key, obj, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 2)
	assert.Len(t, subs, 2)
	assert.Contains(t, subs, converter.Key{
		Kind: "Service-Port",
		Name: "foo/bar/cat/192.2.0.1/17/8000",
	})
	assert.Contains(t, subs, converter.Key{
		Kind: "Service-Port",
		Name: "foo/bar/dog/192.2.0.1/6/200",
	})
}
