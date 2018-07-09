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
	"fmt"
	"strings"

	"k8s.io/client-go/tools/cache"
)

// Key is an identifier of a Translation.
type Key struct {
	Kind      string
	Namespace string
	Name      string

	// Unversioned=true makes this Key and its associated Translation
	// unversioned.  That is, keys and IDs are not changed when
	// TranslationVersion is bumped.  It's necessary for resources
	// which doesn't have IDs in the first place.
	// REVISIT: Probably Key is not the appropriate place to put this info.
	Unversioned bool
}

func newKeyFromClientKey(kind, strKey string) (Key, error) {
	ns, name, err := cache.SplitMetaNamespaceKey(strKey)
	if err != nil {
		return Key{}, err
	}
	return Key{
		Kind:      kind,
		Namespace: ns,
		Name:      name,
	}, nil
}

// Key returns MetaNamespaceKeyFunc style key for the Key
func (k *Key) Key() string {
	if k.Namespace == "" {
		return k.Name
	}
	return fmt.Sprintf("%s/%s", k.Namespace, k.Name)
}

func (k *Key) translationName() string {
	if k.Unversioned {
		// Note: When Unversioned flag was introduced, TranslationVersion
		// happened to be 3.
		return fmt.Sprintf("%s.3.%s", strings.ToLower(k.Kind), k.Name)
	}
	return fmt.Sprintf("%s.%s.%s", strings.ToLower(k.Kind), TranslationVersion, k.Name)
}
