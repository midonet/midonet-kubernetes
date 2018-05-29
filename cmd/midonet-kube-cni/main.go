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
package main

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/containernetworking/cni/pkg/skel"
	cnitypes "github.com/containernetworking/cni/pkg/types"
	cniSpecVersion "github.com/containernetworking/cni/pkg/version"
	"github.com/projectcalico/libcalico-go/lib/logutils"
	"github.com/sirupsen/logrus"
	"github.com/midonet/midonet-kubernetes/pkg/cni/k8s"
	"github.com/midonet/midonet-kubernetes/pkg/cni/types"
	"github.com/midonet/midonet-kubernetes/pkg/cni/utils"
)

func init() {
	// This ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func cmdAdd(args *skel.CmdArgs) error {
	// Unmarshal the network config, and perform validation
	conf := types.NetConf{}
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("failed to load netconf: %v", err)
	}

	utils.ConfigureLogging(conf.LogLevel)

	// Extract WEP identifiers such as pod name, pod namespace (for k8s), containerID, IfName.
	wepIDs, err := utils.GetIdentifiers(args)
	if err != nil {
		return err
	}

	logrus.WithField("EndpointIDs", wepIDs).Info("Extracted identifiers")
	logrus.WithField("NetConfg", conf).Info("Loaded CNI NetConf")

	result, err := k8s.CmdAddK8s(args, conf, *wepIDs)
	if err != nil {
		return err
	}

	// Set Gateway to nil. Calico-IPAM doesn't set it, but host-local does.
	// We modify IPs subnet received from the IPAM plugin (host-local),
	// so Gateway isn't valid anymore. It is also not used anywhere by Calico.
	for _, ip := range result.IPs {
		ip.Gateway = nil
	}

	// Print result to stdout, in the format defined by the requested cniVersion.
	return cnitypes.PrintResult(result, conf.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	conf := types.NetConf{}
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("failed to load netconf: %v", err)
	}

	utils.ConfigureLogging(conf.LogLevel)

	epIDs, err := utils.GetIdentifiers(args)
	if err != nil {
		return err
	}

	logger := logrus.WithFields(logrus.Fields{
		"ContainerID": epIDs.ContainerID,
	})

	return k8s.CmdDelK8s(*epIDs, args, conf, logger)
}

func main() {
	// Set up logging formatting.
	logrus.SetFormatter(&logutils.Formatter{})

	// Install a hook that adds file/line no information.
	logrus.AddHook(&logutils.ContextHook{})

	skel.PluginMain(cmdAdd, cmdDel, cniSpecVersion.All)
}
