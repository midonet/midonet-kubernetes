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
	"net"
	"os"
	"runtime"
	"time"

	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/projectcalico/libcalico-go/lib/logutils"
	log "github.com/sirupsen/logrus"

	k8scni "github.com/midonet/midonet-kubernetes/pkg/cni/k8s"
	"github.com/midonet/midonet-kubernetes/pkg/cni/utils"
	"github.com/midonet/midonet-kubernetes/pkg/converter/node"
	"github.com/midonet/midonet-kubernetes/pkg/k8s"
)

// NOTE(yamamoto): This init function was taken from calico CNI plugin
func init() {
	// This ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func main() {
	// Configure log formatting.
	log.SetFormatter(&logutils.Formatter{})

	// Install a hook that adds file/line no information.
	log.AddHook(&logutils.ContextHook{})

	// Attempt to load configuration.
	config := new(Config)
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
	k8sClientset, _, err := k8s.GetClient(config.Kubeconfig)
	if err != nil {
		log.WithError(err).Fatal("Failed to start")
	}

	nodeName := config.NodeName
	logger := log.WithFields(log.Fields{
		"nodeName": nodeName,
	})
retry:
	podCIDR, err := k8scni.GetNodePodCIDR(k8sClientset, nodeName)
	if err != nil {
		logger.WithError(err).Warn("GetNodePodCIDR")
		time.Sleep(time.Second * 5)
		goto retry
	}
	logger = logger.WithFields(log.Fields{
		"podCIDR": podCIDR,
	})
	si, err := node.GetSubnetInfo(podCIDR)
	if err != nil {
		logger.WithError(err).Fatal("GetSubnetInfo")
	}
	ips := []*current.IPConfig{
		{
			Version: "4",
			Address: si.NodeIP,
			Gateway: si.GatewayIP.IP,
		},
	}
	_, clusterNetwork, err := net.ParseCIDR(config.ClusterCIDR)
	if err != nil {
		logger.WithError(err).Fatal("ClusterCIDR")
	}
	networks := []*net.IPNet{clusterNetwork}
	if config.ServiceCIDR != "" {
		_, serviceNetwork, err := net.ParseCIDR(config.ServiceCIDR)
		if err != nil {
			logger.WithError(err).Fatal("ServiceCIDR")
		}
		networks = append(networks, serviceNetwork)
	}
	contNetNS := utils.GetCurrentThreadNetNSPath()
	contVethName := "midokube-node"
	hostVethName := node.IFName()
	contVethMAC, err := utils.DoNetworking(networks, ips, contNetNS, contVethName, hostVethName, true, logger)
	if err != nil {
		logger.WithError(err).Fatal("DoNetworking")
	}
	logger.WithField("contVethMAC", contVethMAC).Info("Success")

	cniConfigPath := config.CNIConfigPath
	if cniConfigPath != "" {
		file, err := os.OpenFile(cniConfigPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			logger.WithError(err).Fatal("OpenFile")
		}
		err = generateCNIConfig(file, podCIDR)
		if err != nil {
			logger.WithError(err).Fatal("generateCNIConfig")
		}
	}

	// We've done successfully.
	// Loop forever to avoid being restarted by kubernetes.
	// Note that DaemonSet manadates restartPolicy=Always.
	// https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/#pod-template
	for {
		time.Sleep(time.Hour)
	}
}
