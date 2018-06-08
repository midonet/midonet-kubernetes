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
	"fmt"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/midonet"
)

type endpointsConverter struct {
	svcInformer cache.SharedIndexInformer
}

func newEndpointsConverter(svcInformer cache.SharedIndexInformer) converter.Converter {
	return &endpointsConverter{svcInformer}
}

func endpoints(key string, svcIP string, subsets []v1.EndpointSubset) map[string][]endpoint {
	m := make(map[string][]endpoint, 0)
	for _, s := range subsets {
		for _, a := range s.Addresses {
			for _, p := range s.Ports {
				ep := endpoint{
					endpointsKey: key,
					portName:     p.Name,
					svcIP:        svcIP,
					ip:           a.IP,
					port:         int(p.Port),
					protocol:     p.Protocol,
				}
				l := m[p.Name]
				l = append(l, ep)
				m[p.Name] = l
			}
		}
	}
	return m
}

func (c *endpointsConverter) Convert(key converter.Key, obj interface{}, config *midonet.Config) ([]converter.BackendResource, converter.SubResourceMap, error) {
	resources := make([]converter.BackendResource, 0)
	subs := make(converter.SubResourceMap)
	svcObj, exists, err := c.svcInformer.GetIndexer().GetByKey(key.Key())
	if err != nil {
		return nil, nil, err
	}
	if !exists {
		// Ignore Endpoints without the corresponding service.
		// Note: This might or might not be transient.
		return nil, nil, nil
	}
	svcSpec := svcObj.(*v1.Service).Spec
	svcIP := svcSpec.ClusterIP
	if svcSpec.Type != v1.ServiceTypeClusterIP || svcIP == "" || svcIP == v1.ClusterIPNone {
		// Ignore Endpoints without ClusterIP.
		return nil, nil, nil
	}
	endpoint := obj.(*v1.Endpoints)
	for _, eps := range endpoints(key.Key(), svcIP, endpoint.Subsets) {
		for _, ep := range eps {
			// We include almost everything in the key so that a modified
			// endpoint is treated as another resource for the
			// MidoNet side.  Note that MidoNet Chains and Rules are not
			// updateable.
			epKey := converter.Key{
				Kind: "Endpoints", // REVISIT: use a dedicated kind
				Name: fmt.Sprintf("%s/%s/%s/%s/%d/%s", key.Name, ep.portName, svcIP, ep.ip, ep.port, ep.protocol),
			}
			subs[epKey] = &ep
		}
	}
	return resources, subs, nil
}
