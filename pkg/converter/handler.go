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
	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/midonet/midonet-kubernetes/pkg/controller"
)

// SubResource is a pseudo resource to represent a part of a k8s resource.
// For example, we represent a k8s service as a set of "ServicePort"
// sub resources.
type SubResource interface {
	Convert(key Key, config *Config) ([]BackendResource, error)
}

// SubResourceMap represents zero or more SubResources.
// It's returned by a Conveter.Convert().
type SubResourceMap map[Key]SubResource

// Updater stores the converted resources.
type Updater interface {
	// NOTE: Pass GVK explicitly as List'ed objects don't have valid
	// TypeMeta.  https://github.com/kubernetes/kubernetes/issues/3030
	Update(parentKind schema.GroupVersionKind, parentObj interface{}, resources map[Key][]BackendResource) error
}

type converterHandler struct {
	converter Converter
	updater   Updater
	config    *Config
}

// NewHandler creates a converterHandler.
func NewHandler(converter Converter, updater Updater, config *Config) controller.Handler {
	return &converterHandler{
		converter: converter,
		updater:   updater,
		config:    config,
	}
}

func (h *converterHandler) convertSubResources(added SubResourceMap, converted map[Key][]BackendResource, clog *log.Entry) error {
	for k, r := range added {
		v, err := r.Convert(k, h.config)
		if err != nil {
			clog.WithError(err).WithFields(log.Fields{
				"sub-key": k,
			}).Error("failed to convert a sub resource")
			return err
		}
		if len(v) > 0 {
			converted[k] = v
		}
	}
	return nil
}

func (h *converterHandler) Update(strKey string, gvk schema.GroupVersionKind, obj interface{}) error {
	key, err := newKeyFromClientKey(gvk.Kind, strKey)
	if err != nil {
		return err
	}
	converted := make(map[Key][]BackendResource)
	clog := log.WithFields(log.Fields{
		"key": key,
		"obj": obj,
	})
	v, subResources, err := h.converter.Convert(key, obj, h.config)
	if err != nil {
		clog.WithError(err).Error("Failed to convert")
		return err
	}
	if len(v) > 0 {
		converted[key] = v
	}
	err = h.convertSubResources(subResources, converted, clog)
	if err != nil {
		clog.WithError(err).Error("Failed to convert sub resources")
		return err
	}
	err = h.updater.Update(gvk, obj, converted)
	if err != nil {
		clog.WithError(err).Error("Failed to update")
		return err
	}
	return nil
}

func (h *converterHandler) Delete(key string) error {
	log.WithField("key", key).Debug("Delete")
	/* nothing to do */
	return nil
}
