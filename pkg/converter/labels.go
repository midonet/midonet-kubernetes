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

const (
	// We use this label to find Translations for a Kubernetes resource.
	// It's somehow redundant with OwnerReference but you can't use
	// arbitrary fields for field selectors.
	OwnerUIDLabel = "midonet.org/owner-uid"

	// We use this label to find Translations for global resources,
	// that is, Translations which doesn't belong to a particular
	// Kubernetes resource.  (See global.go)
	GlobalLabel = "midonet.org/global"
)
