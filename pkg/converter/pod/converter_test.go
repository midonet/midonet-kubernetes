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

package pod

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/midonet/midonet-kubernetes/pkg/converter"
)

var (
	podAwesome = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				converter.MACAnnotation: "33:22:11:11:22:33",
			},
		},
		Spec: v1.PodSpec{
			NodeName:    "awesome-node",
			HostNetwork: false,
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			PodIP: "10.2.2.2",
		},
	}

	podLessAwesome = &v1.Pod{
		Spec: v1.PodSpec{
			NodeName:    "awesome-node",
			HostNetwork: false,
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			PodIP: "10.2.2.2",
		},
	}

	podWithoutIP = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				converter.MACAnnotation: "33:22:11:11:22:33",
			},
		},
		Spec: v1.PodSpec{
			NodeName:    "awesome-node",
			HostNetwork: false,
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}

	podHostNetwork = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				converter.MACAnnotation: "33:22:11:11:22:33",
			},
		},
		Spec: v1.PodSpec{
			NodeName:    "awesome-node",
			HostNetwork: true,
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			PodIP: "10.2.2.2",
		},
	}

	podSucceeded = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				converter.MACAnnotation: "33:22:11:11:22:33",
			},
		},
		Spec: v1.PodSpec{
			NodeName:    "awesome-node",
			HostNetwork: false,
		},
		Status: v1.PodStatus{
			Phase: v1.PodSucceeded,
			PodIP: "10.2.2.2",
		},
	}

	nodeAwesome = &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				converter.HostIDAnnotation: "44FB4381-C99D-4389-86C3-FDCA765BCBDE",
			},
		},
	}

	nodeLessAwesome = &v1.Node{}
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
		Kind:      "Pod",
		Namespace: "foo",
		Name:      "awesome-pod",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &podConverter{nodeGetter: &objGetter{
		objs: map[string]interface{}{
			"awesome-node": nodeAwesome,
		},
	}}
	rs, subs, err := c.Convert(key, podAwesome, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 2)
	assert.Len(t, subs, 2)
	assert.Contains(t, subs, converter.Key{
		Kind:        "Pod-MAC",
		Name:        "awesome-pod/mac/332211112233",
		Unversioned: true, // XXX
	})
	assert.Contains(t, subs, converter.Key{
		Kind:        "Pod-ARP",
		Name:        "awesome-pod/ip/10.2.2.2/332211112233",
		Unversioned: true, // XXX
	})
}

func TestConverterGetterError(t *testing.T) {
	key := converter.Key{
		Kind:      "Pod",
		Namespace: "foo",
		Name:      "awesome-pod",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &podConverter{nodeGetter: &objErrorGetter{}}
	_, _, err := c.Convert(key, podAwesome, config)
	assert.Error(t, err)
}

func TestConverterNotRunning(t *testing.T) {
	key := converter.Key{
		Kind:      "Pod",
		Namespace: "foo",
		Name:      "awesome-pod",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &podConverter{nodeGetter: &objGetter{
		objs: map[string]interface{}{
			"awesome-node": nodeAwesome,
		},
	}}
	rs, subs, err := c.Convert(key, podSucceeded, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 0)
	assert.Len(t, subs, 0)
}

func TestConverterNoIP(t *testing.T) {
	key := converter.Key{
		Kind:      "Pod",
		Namespace: "foo",
		Name:      "awesome-pod",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &podConverter{nodeGetter: &objGetter{
		objs: map[string]interface{}{
			"awesome-node": nodeAwesome,
		},
	}}
	rs, subs, err := c.Convert(key, podWithoutIP, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 2)
	assert.Len(t, subs, 1)
	assert.Contains(t, subs, converter.Key{
		Kind:        "Pod-MAC",
		Name:        "awesome-pod/mac/332211112233",
		Unversioned: true, // XXX
	})
}

func TestConverterWithoutMACAnnotation(t *testing.T) {
	key := converter.Key{
		Kind:      "Pod",
		Namespace: "foo",
		Name:      "awesome-pod",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &podConverter{nodeGetter: &objGetter{
		objs: map[string]interface{}{
			"awesome-node": nodeAwesome,
		},
	}}
	rs, subs, err := c.Convert(key, podLessAwesome, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 2)
	assert.Len(t, subs, 0)
}

func TestConverterNoNode(t *testing.T) {
	key := converter.Key{
		Kind:      "Pod",
		Namespace: "foo",
		Name:      "awesome-pod",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &podConverter{nodeGetter: &objGetter{}}
	_, _, err := c.Convert(key, podAwesome, config)
	assert.Error(t, err)
}

func TestConverterWithNodeWithoutHostID(t *testing.T) {
	key := converter.Key{
		Kind:      "Pod",
		Namespace: "foo",
		Name:      "awesome-pod",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &podConverter{nodeGetter: &objGetter{
		objs: map[string]interface{}{
			"awesome-node": nodeLessAwesome,
		},
	}}
	_, _, err := c.Convert(key, podAwesome, config)
	assert.Error(t, err)
}

func TestConverterHostNetwork(t *testing.T) {
	key := converter.Key{
		Kind:      "Pod",
		Namespace: "foo",
		Name:      "awesome-pod",
	}
	config := &converter.Config{
		Tenant: "MyTenant",
	}
	c := &podConverter{nodeGetter: &objGetter{
		objs: map[string]interface{}{
			"awesome-node": nodeAwesome,
		},
	}}
	rs, subs, err := c.Convert(key, podHostNetwork, config)
	assert.Nil(t, err)
	assert.Len(t, rs, 0)
	assert.Len(t, subs, 0)
}
