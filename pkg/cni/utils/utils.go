// Copyright 2015 Tigera Inc
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

package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/containernetworking/cni/pkg/skel"
	cnitypes "github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/midonet/midonet-kubernetes/pkg/cni/types"
	"github.com/sirupsen/logrus"
)

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CleanUpIPAM calls IPAM plugin to release the IP address.
// It also contains IPAM plugin specific changes needed before calling the plugin.
func CleanUpIPAM(conf types.NetConf, args *skel.CmdArgs, logger *logrus.Entry) error {
	fmt.Fprint(os.Stderr, "MidoNet CNI releasing IP address\n")
	logger.WithFields(logrus.Fields{"paths": os.Getenv("CNI_PATH"),
		"type": conf.IPAM.Type}).Debug("Looking for IPAM plugin in paths")

	// host-local IPAM releases the IP by ContainerID, so podCidr isn't really used to release the IP.
	// It just needs a valid CIDR, but it doesn't have to be the CIDR associated with the host.
	dummyPodCidr := "0.0.0.0/0"
	var stdinData map[string]interface{}

	err := json.Unmarshal(args.StdinData, &stdinData)
	if err != nil {
		return err
	}

	logger.WithField("podCidr", dummyPodCidr).Info("Using a dummy podCidr to release the IP")
	stdinData["ipam"].(map[string]interface{})["subnet"] = dummyPodCidr

	args.StdinData, err = json.Marshal(stdinData)
	if err != nil {
		return err
	}
	logger.WithField("stdin", string(args.StdinData)).Debug("Updated stdin data for Delete Cmd")

	err = ipam.ExecDel(conf.IPAM.Type, args.StdinData)

	if err != nil {
		logger.Error(err)
	}

	return err
}

type WEPIdentifiers struct {
	ContainerID string
	Endpoint    string
	Namespace   string
	Pod         string
}

// GetIdentifiers takes CNI command arguments, and extracts identifiers i.e. pod name, pod namespace,
// container ID, endpoint(container interface name) and orchestratorID based on the orchestrator.
func GetIdentifiers(args *skel.CmdArgs) (*WEPIdentifiers, error) {
	// Determine if running under k8s by checking the CNI args
	k8sArgs := types.K8sArgs{}
	if err := cnitypes.LoadArgs(args.Args, &k8sArgs); err != nil {
		return nil, err
	}
	logrus.Debugf("Getting WEP identifiers with arguments: %s", args.Args)
	logrus.Debugf("Loaded k8s arguments: %v", k8sArgs)

	epIDs := WEPIdentifiers{}
	epIDs.ContainerID = args.ContainerID
	epIDs.Endpoint = args.IfName

	// Check if the workload is running under Kubernetes.
	if string(k8sArgs.K8S_POD_NAMESPACE) != "" && string(k8sArgs.K8S_POD_NAME) != "" {
		epIDs.Pod = string(k8sArgs.K8S_POD_NAME)
		epIDs.Namespace = string(k8sArgs.K8S_POD_NAMESPACE)
	} else {
		logrus.Fatal("No K8S parameters")
	}

	return &epIDs, nil
}

// ReleaseIPAllocation is called to cleanup IPAM allocations if something goes wrong during
// CNI ADD execution.
func ReleaseIPAllocation(logger *logrus.Entry, ipamType string, stdinData []byte) {
	logger.Info("Cleaning up IP allocations for failed ADD")
	if err := os.Setenv("CNI_COMMAND", "DEL"); err != nil {
		// Failed to set CNI_COMMAND to DEL.
		logger.Warning("Failed to set CNI_COMMAND=DEL")
	} else {
		if err := ipam.ExecDel(ipamType, stdinData); err != nil {
			// Failed to cleanup the IP allocation.
			logger.Warning("Failed to clean up IP allocations for failed ADD")
		}
	}
}

// Set up logging for both Calico and libcalico using the provided log level,
func ConfigureLogging(logLevel string) {
	if strings.EqualFold(logLevel, "debug") {
		logrus.SetLevel(logrus.DebugLevel)
	} else if strings.EqualFold(logLevel, "info") {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		// Default level
		logrus.SetLevel(logrus.WarnLevel)
	}

	logrus.SetOutput(os.Stderr)
}
