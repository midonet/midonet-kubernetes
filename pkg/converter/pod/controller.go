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

package pod

import (
	"k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	mncli "github.com/midonet/midonet-kubernetes/pkg/client/clientset/versioned"
	mninformers "github.com/midonet/midonet-kubernetes/pkg/client/informers/externalversions"
	"github.com/midonet/midonet-kubernetes/pkg/controller"
	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/midonet"
)

func NewController(si informers.SharedInformerFactory, msi mninformers.SharedInformerFactory, kc *kubernetes.Clientset, mc *mncli.Clientset, recorder record.EventRecorder, config *midonet.Config) *controller.Controller {
	informer := si.Core().V1().Pods().Informer()
	nodeInformer := si.Core().V1().Nodes().Informer()
	updater := converter.NewTranslationUpdater(mc)
	handler := converter.NewHandler(newPodConverter(nodeInformer), updater, config)
	gvk := v1.SchemeGroupVersion.WithKind("Pod")
	return controller.NewController(gvk, informer, handler)
}
