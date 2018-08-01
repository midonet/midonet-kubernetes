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

package endpoints

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"

	"github.com/midonet/midonet-kubernetes/pkg/converter"
)

var (
	endpointsCatDog = &v1.Endpoints{
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{IP: "10.0.0.1"},
					{IP: "10.0.0.2"},
				},
				Ports: []v1.EndpointPort{
					{Name: "cat", Port: 18000, Protocol: "UDP"},
					{Name: "dog", Port: 10200, Protocol: "TCP"},
				},
			},
			{
				Addresses: []v1.EndpointAddress{
					{IP: "10.0.0.3"},
					{IP: "10.0.0.4"},
				},
				Ports: []v1.EndpointPort{
					{Name: "dog", Port: 10200, Protocol: "TCP"},
				},
			},
		},
	}

	svcCatDog = &v1.Service{
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
					Name:     "sheep",
					Protocol: "TCP",
					Port:     1000,
				},
				{
					Name:     "dog",
					Protocol: "TCP",
					Port:     200,
				},
			},
		},
	}

	svcEmptyClusterIP = &v1.Service{
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: "",
			Ports: []v1.ServicePort{
				{
					Name:     "cat",
					Protocol: "UDP",
					Port:     8000,
				},
				{
					Name:     "sheep",
					Protocol: "TCP",
					Port:     1000,
				},
				{
					Name:     "dog",
					Protocol: "TCP",
					Port:     200,
				},
			},
		},
	}

	svcClusterIPNone = &v1.Service{
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
					Name:     "sheep",
					Protocol: "TCP",
					Port:     1000,
				},
				{
					Name:     "dog",
					Protocol: "TCP",
					Port:     200,
				},
			},
		},
	}

	svcNoPorts = &v1.Service{
		Spec: v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: "192.2.0.1",
		},
	}
)

type objGetter struct {
	objs map[string]interface{}
}

func (s *objGetter) GetByKey(key string) (interface{}, bool, error) {
	obj, exists := s.objs[key]
	return obj, exists, nil
}

type objErrorGetter struct{}

func (s *objErrorGetter) GetByKey(key string) (interface{}, bool, error) {
	return nil, false, errors.New("Some error")
}

func TestConverter(t *testing.T) {
	key := converter.Key{
		Kind:      "Endpoints",
		Namespace: "foo",
		Name:      "bar",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &endpointsConverter{svcGetter: &objGetter{
		objs: map[string]interface{}{
			"foo/bar": svcCatDog,
		},
	}}
	rs, subs, err := c.Convert(key, endpointsCatDog, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 0)
	assert.Len(t, subs, 6)
	assert.Contains(t, subs, converter.Key{
		Kind: "Endpoints-Port",
		Name: "bar/cat/192.2.0.1/10.0.0.1/18000/UDP",
	})
	assert.Contains(t, subs, converter.Key{
		Kind: "Endpoints-Port",
		Name: "bar/cat/192.2.0.1/10.0.0.2/18000/UDP",
	})
	assert.Contains(t, subs, converter.Key{
		Kind: "Endpoints-Port",
		Name: "bar/dog/192.2.0.1/10.0.0.1/10200/TCP",
	})
	assert.Contains(t, subs, converter.Key{
		Kind: "Endpoints-Port",
		Name: "bar/dog/192.2.0.1/10.0.0.2/10200/TCP",
	})
	assert.Contains(t, subs, converter.Key{
		Kind: "Endpoints-Port",
		Name: "bar/dog/192.2.0.1/10.0.0.3/10200/TCP",
	})
	assert.Contains(t, subs, converter.Key{
		Kind: "Endpoints-Port",
		Name: "bar/dog/192.2.0.1/10.0.0.4/10200/TCP",
	})
}

func TestConverterServiceNoPorts(t *testing.T) {
	key := converter.Key{
		Kind:      "Endpoints",
		Namespace: "foo",
		Name:      "bar",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &endpointsConverter{svcGetter: &objGetter{
		objs: map[string]interface{}{
			"foo/bar": svcNoPorts,
		},
	}}
	rs, subs, err := c.Convert(key, endpointsCatDog, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 0)
	assert.Len(t, subs, 6)
	assert.Contains(t, subs, converter.Key{
		Kind: "Endpoints-Port",
		Name: "bar/cat/192.2.0.1/10.0.0.1/18000/UDP",
	})
	assert.Contains(t, subs, converter.Key{
		Kind: "Endpoints-Port",
		Name: "bar/cat/192.2.0.1/10.0.0.2/18000/UDP",
	})
	assert.Contains(t, subs, converter.Key{
		Kind: "Endpoints-Port",
		Name: "bar/dog/192.2.0.1/10.0.0.1/10200/TCP",
	})
	assert.Contains(t, subs, converter.Key{
		Kind: "Endpoints-Port",
		Name: "bar/dog/192.2.0.1/10.0.0.2/10200/TCP",
	})
	assert.Contains(t, subs, converter.Key{
		Kind: "Endpoints-Port",
		Name: "bar/dog/192.2.0.1/10.0.0.3/10200/TCP",
	})
	assert.Contains(t, subs, converter.Key{
		Kind: "Endpoints-Port",
		Name: "bar/dog/192.2.0.1/10.0.0.4/10200/TCP",
	})
}

func TestConverterEmptyClusterIP(t *testing.T) {
	key := converter.Key{
		Kind:      "Endpoints",
		Namespace: "foo",
		Name:      "bar",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &endpointsConverter{svcGetter: &objGetter{
		objs: map[string]interface{}{
			"foo/bar": svcEmptyClusterIP,
		},
	}}
	rs, subs, err := c.Convert(key, endpointsCatDog, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 0)
	assert.Len(t, subs, 0)
}

func TestConverterClusterIPNone(t *testing.T) {
	key := converter.Key{
		Kind:      "Endpoints",
		Namespace: "foo",
		Name:      "bar",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &endpointsConverter{svcGetter: &objGetter{
		objs: map[string]interface{}{
			"foo/bar": svcClusterIPNone,
		},
	}}
	rs, subs, err := c.Convert(key, endpointsCatDog, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 0)
	assert.Len(t, subs, 0)
}

func TestConverterWithoutService(t *testing.T) {
	key := converter.Key{
		Kind:      "Endpoints",
		Namespace: "foo",
		Name:      "bar",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &endpointsConverter{svcGetter: &objGetter{}}
	rs, subs, err := c.Convert(key, endpointsCatDog, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 0)
	assert.Len(t, subs, 0)
}

func TestConverterGetterError(t *testing.T) {
	key := converter.Key{
		Kind:      "Endpoints",
		Namespace: "foo",
		Name:      "bar",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &endpointsConverter{svcGetter: &objErrorGetter{}}
	_, _, err := c.Convert(key, endpointsCatDog, config)
	assert.Error(t, err)
}
