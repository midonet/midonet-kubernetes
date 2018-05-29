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

package midonet

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/midonet/midonet-kubernetes/pkg/apis/midonet/v1"
)

type APIResource interface {
	Path(string) string
	MediaType() string
}

// This struct is just a marker
type midonetResource struct {
}

func (_ *midonetResource) ToAPI(res interface{}) (*v1.BackendResource, error) {
	data, err := json.Marshal(res)
	if err != nil {
		log.WithError(err).Error("Marshal")
		return nil, err
	}
	r := &v1.BackendResource{
		Kind: TypeNameForObject(res),
		Body: string(data),
	}
	hasparent, ok := res.(HasParent)
	if ok {
		r.Parent = hasparent.GetParent().String()
	}
	return r, nil
}

func FromAPI(res v1.BackendResource) (APIResource, error) {
	r := ObjectByTypeName(res.Kind).(APIResource)
	err := json.Unmarshal([]byte(res.Body), r)
	if err != nil {
		return nil, err
	}
	if res.Parent != "" {
		hasparent, ok := r.(HasParent)
		if !ok {
			return nil, fmt.Errorf("Unexpected Parent for %s", res.Kind)
		}
		id, err := uuid.Parse(res.Parent)
		if err != nil {
			return nil, err
		}
		hasparent.SetParent(&id)
	}
	return r, nil
}
