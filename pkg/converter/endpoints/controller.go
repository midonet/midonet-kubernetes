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

package endpoints

import (
	"k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	"github.com/yamt/midonet-kubernetes/pkg/controller"
	"github.com/yamt/midonet-kubernetes/pkg/converter"
	"github.com/yamt/midonet-kubernetes/pkg/midonet"
	mncli "github.com/yamt/midonet-kubernetes/pkg/client/clientset/versioned"
	mninformers "github.com/yamt/midonet-kubernetes/pkg/client/informers/externalversions"
)

func NewController(si informers.SharedInformerFactory, msi mninformers.SharedInformerFactory, kc *kubernetes.Clientset, mc *mncli.Clientset, config *midonet.Config) *controller.Controller {
	informer := si.Core().V1().Endpoints().Informer()
	svcInformer := si.Core().V1().Services().Informer()
	updater := converter.NewTranslationUpdater(mc)
	handler := converter.NewHandler(newEndpointsConverter(svcInformer), updater, config)
	gvk := v1.SchemeGroupVersion.WithKind("Endpoints")
	c := controller.NewController(gvk, informer, handler)
	// Kick the Endpoints controller when the corresponding Service is updated.
	svcInformer.AddEventHandler(controller.NewEventHandler("svc-eps", c.GetQueue()))
	return c
}
