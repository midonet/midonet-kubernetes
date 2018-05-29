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

package pusher

import (
	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/midonet/midonet-kubernetes/pkg/apis/midonet/v1"
	mncli "github.com/midonet/midonet-kubernetes/pkg/client/clientset/versioned"
	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/midonet"
	"github.com/midonet/midonet-kubernetes/pkg/util"
)

type Handler struct {
	mncli  *mncli.Clientset
	client *midonet.Client
	config *midonet.Config
}

func newHandler(mc *mncli.Clientset, config *midonet.Config) *Handler {
	client := midonet.NewClient(config)
	return &Handler{
		mncli:  mc,
		client: client,
		config: config,
	}
}

func (h *Handler) Update(key string, gvk schema.GroupVersionKind, obj interface{}) error {
	tr := obj.(*v1.Translation)
	clog := log.WithFields(log.Fields{
		"key": key,
	})
	var resources []midonet.APIResource
	for _, res := range tr.Resources {
		r, err := midonet.FromAPI(res)
		if err != nil {
			return err
		}
		resources = append(resources, r)
	}
	if tr.ObjectMeta.DeletionTimestamp == nil {
		clog.Debug("Handling Translation Update")
		err := h.client.Push(resources)
		if err != nil {
			return err
		}
		clog.Info("Translation Update pushed to the backend")
	} else {
		clog.Debug("Handling Translation Deletion")
		err := h.client.Delete(resources)
		if err != nil {
			return err
		}
		clog.Info("Translation Deletion pushed to the backend")
		err = h.clearFinalizer(tr)
		if err != nil {
			return err
		}
		clog.Info("Removed finalizer")
	}
	return nil
}

func (h *Handler) clearFinalizer(tr *v1.Translation) error {
	ns := tr.ObjectMeta.Namespace
	new := tr.DeepCopy()
	finalizers, removed := util.RemoveFirst(new.ObjectMeta.Finalizers, converter.MidoNetAPIDeleter)
	if !removed {
		/* nothing to do */
		return nil
	}
	new.ObjectMeta.Finalizers = finalizers
	_, err := h.mncli.MidonetV1().Translations(ns).Update(new)
	return err
}

func (h *Handler) Delete(key string) error {
	/* nothing to do */
	return nil
}
