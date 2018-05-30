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

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/midonet"
)

func IDForKey(key string) uuid.UUID {
	return converter.IDForKey("Pod", key)
}

type podConverter struct {
	nodeInformer cache.SharedIndexInformer
}

func newPodConverter(nodeInformer cache.SharedIndexInformer) converter.Converter {
	return &podConverter{nodeInformer}
}

func (c *podConverter) Convert(key string, obj interface{}, config *midonet.Config) ([]converter.BackendResource, converter.SubResourceMap, error) {
	clog := log.WithField("key", key)
	baseID := IDForKey(key)
	bridgePortID := baseID
	spec := obj.(*v1.Pod).Spec
	nodeName := spec.NodeName
	if nodeName == "" {
		clog.Info("NodeName is not set")
		return nil, nil, nil
	}
	if spec.HostNetwork {
		clog.Info("hostNetwork")
		return nil, nil, nil
	}
	bridgeID := converter.IDForKey("Node", nodeName)
	nodeObj, exists, err := c.nodeInformer.GetIndexer().GetByKey(nodeName)
	if err != nil {
		return nil, nil, err
	}
	if !exists {
		return nil, nil, fmt.Errorf("Node %s is not known yet.", nodeName)
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
			InterfaceName: IFNameForKey(key),
		},
	}
	return res, nil, nil
}
