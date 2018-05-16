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

	"github.com/yamt/midonet-kubernetes/pkg/midonet"
)

// SubResource is a pseudo resource to represent a part of a k8s resource.
// For example, we represent a k8s service as a set of "ServicePort"
// sub resources.
type SubResource interface {
	Convert(key string, config *midonet.Config) ([]BackendResource, error)
}

type SubResourceMap map[string]SubResource

type Updater interface {
	// NOTE: Pass GVK explicitly as List'ed objects don't have valid
	// TypeMeta.  https://github.com/kubernetes/kubernetes/issues/3030
	Update(key string, parentKind schema.GroupVersionKind, parentObj interface{}, resources map[string][]BackendResource) error
	Delete(key string) error
}

type Handler struct {
	converter Converter
	updater   Updater
	config    *midonet.Config

	resolver *midonet.HostResolver
}

func NewHandler(converter Converter, updater Updater, config *midonet.Config) *Handler {
	client := midonet.NewClient(config)
	return &Handler{
		converter: converter,
		updater:   updater,
		config:    config,
		resolver:  midonet.NewHostResolver(client),
	}
}

func (h *Handler) convertSubResources(key string, parentObj interface{}, added SubResourceMap, converted map[string][]BackendResource, clog *log.Entry) error {
	for k, r := range added {
		v, err := r.Convert(k, h.config)
		if err != nil {
			clog.WithError(err).WithFields(log.Fields{
				"key":     key,
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

func (h *Handler) Update(key string, gvk schema.GroupVersionKind, obj interface{}) error {
	converted := make(map[string][]BackendResource)
	clog := log.WithFields(log.Fields{
		"key": key,
		"obj": obj,
	})
	v, subResources, err := h.converter.Convert(key, obj, h.config, h.resolver)
	if err != nil {
		clog.WithError(err).Error("Failed to convert")
		return err
	}
	if len(v) > 0 {
		converted[key] = v
	}
	err = h.convertSubResources(key, obj, subResources, converted, clog)
	if err != nil {
		clog.WithError(err).Error("Failed to convert sub resources")
		return err
	}
	err = h.updater.Update(key, gvk, obj, converted)
	if err != nil {
		clog.WithError(err).Error("Failed to update")
		return err
	}
	return nil
}

func (h *Handler) Delete(key string) error {
	clog := log.WithField("key", key)
	err := h.updater.Delete(key)
	if err != nil {
		clog.WithError(err).Error("Failed to delete")
		return err
	}
	return nil
}
