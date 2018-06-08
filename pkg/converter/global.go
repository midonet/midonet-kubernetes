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
	"github.com/google/uuid"

	mncli "github.com/midonet/midonet-kubernetes/pkg/client/clientset/versioned"
	"github.com/midonet/midonet-kubernetes/pkg/midonet"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func ServicesChainID(config *midonet.Config) uuid.UUID {
	baseID := IDForTenant(config.Tenant)
	return SubID(baseID, "Services Chain")
}

func MainChainID(config *midonet.Config) uuid.UUID {
	return IDForTenant(config.Tenant)
}

func GlobalResources(config *midonet.Config) map[Key]([]BackendResource) {
	tenant := config.Tenant
	baseID := IDForTenant(tenant)
	mainChainID := baseID
	preChainID := SubID(baseID, "Pre Chain")
	servicesChainID := ServicesChainID(config)
	jumpToPreRuleID := SubID(baseID, "Jump To Pre")
	jumpToServicesRuleID := SubID(baseID, "Jump To Services")
	revSNATRuleID := SubID(baseID, "Reverse SNAT")
	revDNATRuleID := SubID(baseID, "Reverse DNAT")
	return map[Key]([]BackendResource){
		// Chains shared among Bridges for Nodes
		Key{Kind: "midonet-global", Name: "chain"}: []BackendResource{
			&midonet.Chain{
				ID:       &mainChainID,
				Name:     "KUBE-MAIN",
				TenantID: tenant,
			},
			&midonet.Chain{
				ID:       &preChainID,
				Name:     "KUBE-PRE",
				TenantID: tenant,
			},
			&midonet.Chain{
				ID:       &servicesChainID,
				Name:     "KUBE-SERVICES",
				TenantID: tenant,
			},
			midonet.JumpRule(&jumpToServicesRuleID, &mainChainID, &servicesChainID),
			midonet.JumpRule(&jumpToPreRuleID, &mainChainID, &preChainID),
			// Reverse NAT rules for Endpoints.
			&midonet.Rule{
				Parent:     midonet.Parent{ID: &preChainID},
				ID:         &revDNATRuleID,
				Type:       "rev_dnat",
				FlowAction: "accept",
			},
			&midonet.Rule{
				Parent:     midonet.Parent{ID: &preChainID},
				ID:         &revSNATRuleID,
				Type:       "rev_snat",
				FlowAction: "continue",
			},
		},
	}
}

func EnsureGlobalResources(mncli mncli.Interface, config *midonet.Config) error {
	resources := GlobalResources(config)
	updater := NewTranslationUpdater(mncli)
	return updater.Update(schema.GroupVersionKind{}, nil, resources)
}
