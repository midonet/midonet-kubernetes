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

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"

	mnv1 "github.com/midonet/midonet-kubernetes/pkg/apis/midonet/v1"
	mncli "github.com/midonet/midonet-kubernetes/pkg/client/clientset/versioned"
	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/midonet"
	"github.com/midonet/midonet-kubernetes/pkg/util"
)

type pusherHandler struct {
	mncli    *mncli.Clientset
	client   *midonet.Client
	recorder record.EventRecorder
	config   *midonet.Config
}

func newHandler(mc *mncli.Clientset, recorder record.EventRecorder, config *midonet.Config) *pusherHandler {
	client := midonet.NewClient(config)
	return &pusherHandler{
		mncli:    mc,
		client:   client,
		recorder: recorder,
		config:   config,
	}
}

func (h *pusherHandler) Update(key string, gvk schema.GroupVersionKind, obj interface{}) error {
	tr := obj.(*mnv1.Translation)
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
			h.recorder.Eventf(tr, v1.EventTypeWarning, "TranslationUpdateError", "Translation Update failed with error %v", err)
			return err
		}
		h.recorder.Event(tr, v1.EventTypeNormal, "TranslationUpdatePushed", "Translation Update pushed to the backend")
	} else {
		clog.Debug("Handling Translation Deletion")
		err := h.client.Delete(resources)
		if err != nil {
			h.recorder.Eventf(tr, v1.EventTypeWarning, "TranslationDeletionError", "Translation Deletion failed with error %v", err)
			return err
		}
		h.recorder.Event(tr, v1.EventTypeNormal, "TranslationDeletionPushed", "Translation Deletion pushed to the backend")
		err = h.clearFinalizer(tr)
		if err != nil {
			h.recorder.Eventf(tr, v1.EventTypeWarning, "TranslationDeletionFinalizerError", "Translation Deletion failed to remove finalizer with error %v", err)
			return err
		}
		clog.Info("Removed finalizer")
	}
	return nil
}

func (h *pusherHandler) clearFinalizer(tr *mnv1.Translation) error {
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

func (h *pusherHandler) Delete(key string) error {
	/* nothing to do */
	return nil
}
