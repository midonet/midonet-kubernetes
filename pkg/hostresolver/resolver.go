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

package hostresolver

import (
	log "github.com/sirupsen/logrus"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/midonet"
)

type Handler struct {
	kc       *kubernetes.Clientset
	resolver *midonet.HostResolver
	recorder record.EventRecorder
	config   *midonet.Config
}

func newHandler(kc *kubernetes.Clientset, recorder record.EventRecorder, config *midonet.Config) *Handler {
	client := midonet.NewClient(config)
	resolver := midonet.NewHostResolver(client)
	return &Handler{
		kc:       kc,
		resolver: resolver,
		recorder: recorder,
		config:   config,
	}
}

func (h *Handler) Update(key string, gvk schema.GroupVersionKind, obj interface{}) error {
	n := obj.(*v1.Node)
	new := n.DeepCopy()
	annotations := new.ObjectMeta.Annotations
	clog := log.WithFields(log.Fields{
		"node": key,
	})
	clog.Debug("HostResolver Node update handler")
	_, ok := annotations[converter.HostIDAnnotation]
	if ok {
		/* nothing to do */
		return nil
	}
	id, err := h.resolver.ResolveHost(key)
	if err != nil {
		return err
	}
	annotations[converter.HostIDAnnotation] = id.String()
	// REVISIT: should use Patch to avoid clearing unknown fields
	_, err = h.kc.CoreV1().Nodes().Update(new)
	if err != nil {
		return err
	}
	h.recorder.Eventf(n, v1.EventTypeNormal, "MidoNetHostIDAnnotated", "Annotated with MidoNet Host ID %s", id.String())
	return nil
}

func (h *Handler) Delete(key string) error {
	/* nothing to do */
	return nil
}
