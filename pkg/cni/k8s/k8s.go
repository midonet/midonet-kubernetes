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

package k8s

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/midonet/midonet-kubernetes/pkg/cni/types"
	"github.com/midonet/midonet-kubernetes/pkg/cni/utils"
	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/converter/node"
	"github.com/midonet/midonet-kubernetes/pkg/converter/pod"
	nodecli "github.com/midonet/midonet-kubernetes/pkg/nodeapi/client"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// CmdAddK8s performs the "ADD" operation on a kubernetes pod
// Having kubernetes code in its own file avoids polluting the mainline code. It's expected that the kubernetes case will
// more special casing than the mainline code.
func CmdAddK8s(args *skel.CmdArgs, conf types.NetConf, epIDs utils.WEPIdentifiers) (*current.Result, error) {
	var err error
	var result *current.Result

	utils.ConfigureLogging(conf.LogLevel)

	logger := logrus.WithFields(logrus.Fields{
		"ContainerID": epIDs.ContainerID,
		"Pod":         epIDs.Pod,
		"Namespace":   epIDs.Namespace,
	})

	logger.Info("Extracted identifiers for CmdAddK8s")

	podCIDR := conf.Kubernetes.PodCIDR
	if podCIDR == "" {
		fmt.Fprint(os.Stderr, "MidoNet CNI fetching podCidr from Kubernetes\n")

		client, err := newK8sClient(conf, logger)
		if err != nil {
			return nil, err
		}
		logger.WithField("client", client).Debug("Created Kubernetes client")

		cidr, err := getPodCidr(client, conf, epIDs.Namespace, epIDs.Pod)
		if err != nil {
			logger.Info("Failed to getPodCidr")
			return nil, err
		}
		logger.WithField("podCidr", cidr).Info("Fetched podCidr")
		podCIDR = cidr
	}

	subnetInfo, err := node.GetSubnetInfo(podCIDR)
	if err != nil {
		return nil, err
	}
	gatewayIP := subnetInfo.GatewayIP
	nodeIP := subnetInfo.NodeIP

	// Replace the actual value in the args.StdinData as that's what's passed to the IPAM plugin.
	var stdinData map[string]interface{}
	if err := json.Unmarshal(args.StdinData, &stdinData); err != nil {
		return nil, err
	}
	stdinData["ipam"].(map[string]interface{})["subnet"] = podCIDR
	stdinData["ipam"].(map[string]interface{})["gateway"] = gatewayIP.IP.String()
	fmt.Fprintf(os.Stderr, "MidoNet CNI passing podCidr to host-local IPAM: %s\n", podCIDR)
	args.StdinData, err = json.Marshal(stdinData)
	if err != nil {
		return nil, err
	}
	logger.WithField("stdin", string(args.StdinData)).Debug("Updated stdin data")

retry_ipam:
	logger.Debugf("Calling IPAM plugin %s", conf.IPAM.Type)
	ipamResult, err := ipam.ExecAdd(conf.IPAM.Type, args.StdinData)
	if err != nil {
		return nil, err
	}
	logger.Debugf("IPAM plugin returned: %+v", ipamResult)

	// Convert IPAM result into current Result.
	// IPAM result has a bunch of fields that are optional for an IPAM plugin
	// but required for a CNI plugin, so this is to populate those fields.
	// See CNI Spec doc for more details.
	result, err = current.NewResultFromResult(ipamResult)
	if err != nil {
		utils.ReleaseIPAllocation(logger, conf.IPAM.Type, args.StdinData)
		return nil, err
	}

	if len(result.IPs) == 0 {
		utils.ReleaseIPAllocation(logger, conf.IPAM.Type, args.StdinData)
		return nil, errors.New("IPAM plugin returned missing IP config")
	}

	for _, ip := range result.IPs {
		if ip.Address.IP.Equal(nodeIP.IP) {
			// Just leak it and retry
			goto retry_ipam
		}
	}

	// maybeReleaseIPAM cleans up any IPAM allocations if we were creating a new endpoint;
	// it is a no-op if this was a re-network of an existing endpoint.
	maybeReleaseIPAM := func() {
		logger.Debug("Checking if we need to clean up IPAM.")
		//		logger := logger.WithField("IPs", endpoint.Spec.IPNetworks)
		logger.Info("Releasing IPAM allocation after failure")
		utils.ReleaseIPAllocation(logger, conf.IPAM.Type, args.StdinData)
	}

	// Whether the endpoint existed or not, the veth needs (re)creating.
	_, defaultNetwork, _ := net.ParseCIDR("0.0.0.0/0")
	destNetworks := []*net.IPNet{defaultNetwork}
	podKey := fmt.Sprintf("%s/%s", epIDs.Namespace, epIDs.Pod)
	hostVethName := pod.IFNameForKey(podKey)
	contVethMac, err := utils.DoNetworking(destNetworks, result.IPs, args.Netns, args.IfName, hostVethName, false, logger)
	if err != nil {
		logger.WithError(err).Error("Error setting up networking")
		maybeReleaseIPAM()
		return nil, err
	}

	mac, err := net.ParseMAC(contVethMac)
	if err != nil {
		logger.WithError(err).WithField("mac", mac).Error("Error parsing container MAC")
		maybeReleaseIPAM()
		return nil, err
	}

	// REVISIT(yamamoto): We've just set up a veth pair. The rest of
	// the plumbing will be done by the controller and the backend
	// asynchronously.  That is, the controller will create necessary
	// MidoNet objects including HostInterfacePort and the MidoNet agent
	// on this node will notice it and actually connect the interface to
	// its datapath.
	// It might be better for us to ensure those asynchronous plumbing is
	// done here.  Otherwise, if the pod is quick enough, it will see
	// the network not available yet.
	// On the other hand, Calico CNI doesn't seem to wait here.  They
	// just create an Endpoint object without waiting for it to be
	// processed.  So it might be ok practically.
	// If we decided to wait, we can do it by watching the interface to
	// see it to be connected to the "midonet" datapath.

	// Try to annotate the Pod with MAC address info
	// Note: The annotation is merely an optimization.
	err = nodecli.AddPodAnnotation(epIDs.Namespace, epIDs.Pod, converter.MACAnnotation, mac.String())
	if err != nil {
		logger.WithError(err).WithField("mac", mac).Error("Failed to annotate Pod with MAC")
	}

	return result, nil
}

// CmdDelK8s performs the "DEL" operation on a kubernetes pod.
// The following logic only applies to kubernetes since it sends multiple DELs for the same
// endpoint. See: https://github.com/kubernetes/kubernetes/issues/44100
func CmdDelK8s(epIDs utils.WEPIdentifiers, args *skel.CmdArgs, conf types.NetConf, logger *logrus.Entry) error {

	// Release the IP address by calling the configured IPAM plugin.
	ipamErr := utils.CleanUpIPAM(conf, args, logger)

	// Clean up namespace by removing the interfaces.
	err := utils.CleanUpNamespace(args, logger)
	if err != nil {
		return err
	}

	// Return the IPAM error if there was one. The IPAM error will be lost if there was also an error in cleaning up
	// the device or endpoint, but crucially, the user will know the overall operation failed.
	if ipamErr != nil {
		return ipamErr
	}

	logger.Info("Teardown processing complete.")

	return nil
}

func newK8sClient(conf types.NetConf, logger *logrus.Entry) (*kubernetes.Clientset, error) {
	// Some config can be passed in a kubeconfig file
	kubeconfig := conf.Kubernetes.Kubeconfig

	// Config can be overridden by config passed in explicitly in the network config.
	configOverrides := &clientcmd.ConfigOverrides{}

	// Also allow the K8sAPIRoot to appear under the "kubernetes" block in the network config.
	if conf.Kubernetes.K8sAPIRoot != "" {
		configOverrides.ClusterInfo.Server = conf.Kubernetes.K8sAPIRoot
	}

	// Use the kubernetes client code to load the kubeconfig file and combine it with the overrides.
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		configOverrides).ClientConfig()
	if err != nil {
		return nil, err
	}

	logger.Debugf("Kubernetes config %v", config)

	// Create the clientset
	return kubernetes.NewForConfig(config)
}

func getPodCidr(client *kubernetes.Clientset, conf types.NetConf, podNamespace string, podName string) (string, error) {
	// Pull the node name out of the config if it's set. Defaults to nodename
	var nodename string
	if conf.Kubernetes.NodeName != "" {
		nodename = conf.Kubernetes.NodeName
	} else {
		pod, err := client.CoreV1().Pods(podNamespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		nodename = pod.Spec.NodeName
	}
	return GetNodePodCIDR(client, nodename)
}

// GetNodePodCIDR queries podCIDR of the node.
func GetNodePodCIDR(client *kubernetes.Clientset, nodename string) (string, error) {
	node, err := client.CoreV1().Nodes().Get(nodename, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if node.Spec.PodCIDR == "" {
		return "", fmt.Errorf("no podCidr for node %s", nodename)
	}
	return node.Spec.PodCIDR, nil
}
