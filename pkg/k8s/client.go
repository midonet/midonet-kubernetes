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

package k8s

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	midonet "github.com/midonet/midonet-kubernetes/pkg/client/clientset/versioned"
)

// GetClients builds and returns Kubernetes client.
func GetClient(kubeconfig string) (*kubernetes.Clientset, *midonet.Clientset, error) {

	// Now build the Kubernetes client, we support in-cluster config and kubeconfig
	// as means of configuring the client.
	k8sconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build kubernetes client config: %s", err)
	}

	// Get Kubernetes clientset
	k8sClientset, err := kubernetes.NewForConfig(k8sconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build kubernetes client: %s", err)
	}

	// Get MidoNet CR clientset
	mnClientset, err := midonet.NewForConfig(k8sconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build MidoNet CR client: %s", err)
	}

	return k8sClientset, mnClientset, nil
}
