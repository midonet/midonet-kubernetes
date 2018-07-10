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
	"fmt"
	"net"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/midonet"
)

func idForKey(key string) uuid.UUID {
	return converter.IDForKey("Pod", key)
}

type podConverter struct {
	nodeInformer cache.SharedIndexInformer
}

func newPodConverter(nodeInformer cache.SharedIndexInformer) converter.Converter {
	return &podConverter{nodeInformer}
}

func (c *podConverter) Convert(key converter.Key, obj interface{}, config *converter.Config) ([]converter.BackendResource, converter.SubResourceMap, error) {
	subs := make(converter.SubResourceMap)
	clog := log.WithField("key", key)
	baseID := idForKey(key.Key())
	bridgePortID := baseID
	spec := obj.(*v1.Pod).Spec
	meta := obj.(*v1.Pod).ObjectMeta
	status := obj.(*v1.Pod).Status
	nodeName := spec.NodeName
	if nodeName == "" {
		clog.Info("NodeName is not set")
		return nil, nil, nil
	}
	if spec.HostNetwork {
		clog.Debug("hostNetwork")
		return nil, nil, nil
	}
	if status.Phase == v1.PodSucceeded || status.Phase == v1.PodFailed {
		clog.Debug("Terminated pod")
		return nil, nil, nil
	}
	bridgeID := converter.IDForKey("Node", nodeName)
	nodeObj, exists, err := c.nodeInformer.GetIndexer().GetByKey(nodeName)
	if err != nil {
		return nil, nil, err
	}
	if !exists {
		return nil, nil, fmt.Errorf("node %s is not known yet", nodeName)
	}
	node := nodeObj.(*v1.Node)
	hostID, err := uuid.Parse(node.ObjectMeta.Annotations[converter.HostIDAnnotation])
	if err != nil {
		// Retry later.  Note: we don't listen Node events.
		return nil, nil, err
	}
	res := []converter.BackendResource{
		&midonet.Port{
			Parent: midonet.Parent{ID: &bridgeID},
			ID:     &bridgePortID,
			Type:   "Bridge",
		},
		&midonet.HostInterfacePort{
			Parent:        midonet.Parent{ID: &hostID},
			HostID:        &hostID,
			PortID:        &bridgePortID,
			InterfaceName: IFNameForKey(key.Key()),
		},
	}
	macStr, exists := meta.Annotations[converter.MACAnnotation]
	if exists {
		mac, err := net.ParseMAC(macStr)
		if err != nil {
			return nil, nil, err
		}
		skey := converter.Key{
			Kind:        "Pod-MAC",
			Name:        fmt.Sprintf("%s/mac/%s", key.Name, DNSifyMAC(mac)),
			Unversioned: true,
		}
		subs[skey] = &PortMAC{
			BridgeID: bridgeID,
			PortID:   bridgePortID,
			MAC:      mac,
		}
		ip := net.ParseIP(status.PodIP)
		if ip != nil {
			skey := converter.Key{
				Kind:        "Pod-ARP",
				Name:        fmt.Sprintf("%s/ip/%s/%s", key.Name, ip, DNSifyMAC(mac)),
				Unversioned: true,
			}
			subs[skey] = &PortARP{
				BridgeID: bridgeID,
				IP:       ip,
				MAC:      mac,
			}
		}
	}
	return res, subs, nil
}
