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
	"fmt"
	"net"

	"k8s.io/api/core/v1"

	log "github.com/sirupsen/logrus"
)

type tunnelEndpointIPAnnotator struct {
}

func (a *tunnelEndpointIPAnnotator) getData(n *v1.Node) (string, error) {
	for _, addr := range n.Status.Addresses {
		typ := addr.Type
		// REVISIT: How about ExternalIP?
		if typ != v1.NodeInternalIP {
			continue
		}
		ip := net.ParseIP(addr.Address)
		if ip == nil {
			// REVISIT: can this happen?
			log.WithFields(log.Fields{
				"node":    n.ObjectMeta.Name,
				"address": addr.Address,
			}).Fatal("Unparsable Node Address")
		}
		return ip.String(), nil
	}
	return "", fmt.Errorf("Node %s has no usable IP for tunnel endpoint IP", n.ObjectMeta.Name)
}
