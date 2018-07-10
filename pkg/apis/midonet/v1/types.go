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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Translation is an ordered set of BackendResources.
// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Translation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Resources []BackendResource `json:"resources"`

	// REVISIT: We can have a status field to track dirty/in-sync status
	// of the Translation with regard to the backend.  That way we can
	// reduce the amount of backend calls on a reboot of the pusher
	// controller.  On the other hand, it would increase the number of
	// RPCs in the normal operations to update the field.
}

// TranslationList is a list of Translations.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TranslationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Translation `json:"items"`
}

// BackendResource describes a MidoNet API resource.
type BackendResource struct {
	Kind   string `json:"kind"`   // "Bridge", "Port", ...
	Parent string `json:"parent"` // UUID
	Body   string `json:"body"`   // JSON string
}
