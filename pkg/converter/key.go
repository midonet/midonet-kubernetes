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

type Key struct {
	Kind      string
	Namespace string
	Name      string
}

func NewKeyFromClientKey(kind, strKey string) (Key, error) {
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

// Returns MetaNamespaceKeyFunc style key
func (k *Key) Key() string {
	if k.Namespace == "" {
		return k.Name
	}
	return fmt.Sprintf("%s/%s", k.Namespace, k.Name)
}

func (k *Key) TranslationName() string {
	// REVISIT: include a version in the name for safer upgrade.
	return fmt.Sprintf("%s.%s", strings.ToLower(k.Kind), k.Name)
}
