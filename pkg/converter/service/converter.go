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
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"

	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/midonet"
)

type serviceConverter struct{}

func newServiceConverter() converter.Converter {
	return &serviceConverter{}
}

func (*serviceConverter) Convert(key converter.Key, obj interface{}, config *converter.Config) ([]converter.BackendResource, converter.SubResourceMap, error) {
	resources := make([]converter.BackendResource, 0)
	subs := make(converter.SubResourceMap)
	spec := obj.(*v1.Service).Spec
	svcIP := spec.ClusterIP
	if spec.Type != v1.ServiceTypeClusterIP || svcIP == "" || svcIP == v1.ClusterIPNone {
		return resources, nil, nil
	}
	for _, p := range spec.Ports {
		// Note: portKey format should be consistent with the
		// endpoints converter so that it can find the right chain
		// to add rules. (portChainID)
		// NameSpace/Name/ServicePort.Name
		portKey := fmt.Sprintf("%s/%s", key.Key(), p.Name)
		portChainID := converter.IDForKey("ServicePort", portKey)
		resources = append(resources, &midonet.Chain{
			ID:       &portChainID,
			Name:     fmt.Sprintf("KUBE-SVC-%s", portKey),
			TenantID: config.Tenant,
		})

		var proto int
		switch p.Protocol {
		case "TCP":
			proto = 6
		case "UDP":
			proto = 17
		default:
			log.WithField("protocol", p.Protocol).Fatal("Unknown protocol")
		}
		port := int(p.Port)
		// Use a separate key for sub resource so that those will
		// be deleted and re-created whenever they got changed.
		// Note that MidoNet Rules are not updateable.
		// The above KUBE-SVC- chain is not a part of this sub resource
		// because we want to avoid re-creating the chain itself
		// as it would remove rules in the chain.  (Those rules are
		// managed by a separate "endpoints" controller.)
		k := converter.Key{
			Kind: "Service-Port",
			Name: fmt.Sprintf("%s/%s/%d/%d", portKey, svcIP, proto, port),
		}
		subs[k] = &servicePort{portKey, svcIP, proto, port}
	}
	return resources, subs, nil
}
