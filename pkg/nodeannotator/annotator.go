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

package nodeannotator

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/k8s"
	"github.com/midonet/midonet-kubernetes/pkg/midonet"
)

type annotator interface {
	getData(*v1.Node) (string, error)
}

type annotatorHandler struct {
	kc         *kubernetes.Clientset
	recorder   record.EventRecorder
	config     *midonet.Config
	annotators map[string]annotator
}

func newHandler(kc *kubernetes.Clientset, recorder record.EventRecorder, config *midonet.Config) *annotatorHandler {
	client := midonet.NewClient(config)
	resolver := midonet.NewHostResolver(client)
	return &annotatorHandler{
		kc:       kc,
		recorder: recorder,
		config:   config,
		annotators: map[string]annotator{
			converter.HostIDAnnotation: &hostIDAnnotator{
				resolver: resolver,
			},
			converter.TunnelZoneIDAnnotation:     &defaultTunnelZoneAnnotator{},
			converter.TunnelEndpointIPAnnotation: &tunnelEndpointIPAnnotator{},
		},
	}
}

func (h *annotatorHandler) Update(key string, gvk schema.GroupVersionKind, obj interface{}) error {
	n := obj.(*v1.Node)
	new := n.DeepCopy()
	if new.ObjectMeta.Annotations == nil {
		new.ObjectMeta.Annotations = make(map[string]string)
	}
	annotations := new.ObjectMeta.Annotations
	clog := log.WithFields(log.Fields{
		"node": key,
	})
	clog.Debug("nodeannotator Node update handler")
	newAnnotations := make(map[string]string)
	for k, a := range h.annotators {
		_, ok := annotations[k]
		if ok {
			/* nothing to do */
			continue
		}
		data, err := a.getData(n)
		if err != nil {
			return err
		}
		newAnnotations[k] = data
	}
	if len(newAnnotations) == 0 {
		return nil
	}
	for k, data := range newAnnotations {
		annotations[k] = data
	}
	oldData, err := json.Marshal(n)
	if err != nil {
		return err
	}
	newData, err := json.Marshal(new)
	if err != nil {
		return err
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, v1.Node{})
	if err != nil {
		return err
	}
	_, err = h.kc.CoreV1().Nodes().Patch(key, types.StrategicMergePatchType, patchBytes)
	if err != nil {
		return err
	}
	ref, err := k8s.GetReferenceForEvent(n)
	if err != nil {
		return err
	}
	h.recorder.Eventf(ref, v1.EventTypeNormal, "MidoNetHostIDAnnotated", "Annotated %v", newAnnotations)
	return nil
}

func (h *annotatorHandler) Delete(key string) error {
	/* nothing to do */
	return nil
}
