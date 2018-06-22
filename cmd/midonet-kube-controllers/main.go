// Copyright (c) 2017 Tigera, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"strings"

	"github.com/projectcalico/libcalico-go/lib/logutils"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"

	mnscheme "github.com/midonet/midonet-kubernetes/pkg/client/clientset/versioned/scheme"
	mninformers "github.com/midonet/midonet-kubernetes/pkg/client/informers/externalversions"
	"github.com/midonet/midonet-kubernetes/pkg/config"
	"github.com/midonet/midonet-kubernetes/pkg/controller"
	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/converter/endpoints"
	"github.com/midonet/midonet-kubernetes/pkg/converter/node"
	"github.com/midonet/midonet-kubernetes/pkg/converter/pod"
	"github.com/midonet/midonet-kubernetes/pkg/converter/service"
	"github.com/midonet/midonet-kubernetes/pkg/k8s"
	"github.com/midonet/midonet-kubernetes/pkg/midonet"
	"github.com/midonet/midonet-kubernetes/pkg/nodeannotator"
	"github.com/midonet/midonet-kubernetes/pkg/pusher"
)

func main() {
	// Configure log formatting.
	log.SetFormatter(&logutils.Formatter{})

	// Install a hook that adds file/line no information.
	log.AddHook(&logutils.ContextHook{})

	// Attempt to load configuration.
	config := new(config.Config)
	if err := config.Parse(); err != nil {
		log.WithError(err).Fatal("Failed to parse config")
	}
	log.WithField("config", config).Info("Loaded configuration from environment")

	// Set the log level based on the loaded configuration.
	logLevel, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)

	// Build clients to be used by the controllers.
	k8sClientset, mnClientset, err := k8s.GetClient(config.Kubeconfig)
	if err != nil {
		log.WithError(err).Fatal("Failed to start")
	}

	converterCfg := converter.NewConfigFromEnvConfig(config)
	midonetCfg := midonet.NewConfigFromEnvConfig(config)

	mnscheme.AddToScheme(scheme.Scheme)
	broadcaster := record.NewBroadcaster()
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "midonet-kube-controllers"})

	broadcaster.StartLogging(func(format string, args ...interface{}) {
		log.Info(fmt.Sprintf(format, args...))
	})
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: k8sClientset.CoreV1().Events("")})

	err = converter.EnsureGlobalResources(mnClientset, converterCfg, recorder)
	if err != nil {
		log.WithError(err).Fatal("EnsureGlobalResources")
	}

	stop := make(chan struct{})
	defer close(stop)

	si := informers.NewSharedInformerFactory(k8sClientset, 0)
	msi := mninformers.NewSharedInformerFactory(mnClientset, 0)
	controllers := make([]*controller.Controller, 0)
	for _, controllerType := range strings.Split(config.EnabledControllers, ",") {
		newController := node.NewController // Just for type inference
		switch controllerType {
		case "node":
			newController = node.NewController
		case "pod":
			newController = pod.NewController
		case "service":
			newController = service.NewController
		case "endpoints":
			newController = endpoints.NewController
		case "pusher":
			newController = pusher.NewController
		case "nodeannotator":
			newController = nodeannotator.NewController
		}
		c := newController(si, msi, k8sClientset, mnClientset, recorder, converterCfg, midonetCfg)
		controllers = append(controllers, c)
	}

	log.Info("Starting the shared informer")
	si.Start(stop)
	si.WaitForCacheSync(stop)
	log.Info("Cache synced")
	msi.Start(stop)
	msi.WaitForCacheSync(stop)
	log.Info("Translation Cache synced")

	for _, c := range controllers {
		go c.Run()
	}

	// Wait forever.
	serveMetrics()
}
