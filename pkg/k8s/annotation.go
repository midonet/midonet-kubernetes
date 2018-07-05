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

package k8s

import (
	"encoding/json"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
)

func makeStrategicMergePatch(old, new, dataStruct interface{}) ([]byte, error) {
	oldData, err := json.Marshal(old)
	if err != nil {
		return nil, err
	}
	newData, err := json.Marshal(new)
	if err != nil {
		return nil, err
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, dataStruct)
	if err != nil {
		return nil, err
	}
	return patchBytes, nil
}

func AddPodAnnotation(client *kubernetes.Clientset, namespace, name, key, value string) error {
	old, err := client.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	new := old.DeepCopy()
	if new.ObjectMeta.Annotations == nil {
		new.ObjectMeta.Annotations = make(map[string]string)
	}
	new.ObjectMeta.Annotations[key] = value
	patchBytes, err := makeStrategicMergePatch(old, new, v1.Pod{})
	if err != nil {
		return err
	}
	_, err = client.CoreV1().Pods(namespace).Patch(name, types.StrategicMergePatchType, patchBytes)
	// REVISIT: maybe worth a retry in case of version mismatch?
	return err
}

func DeletePodAnnotation(client *kubernetes.Clientset, namespace, name, key string) error {
	old, err := client.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	new := old.DeepCopy()
	delete(new.ObjectMeta.Annotations, key)
	if len(new.ObjectMeta.Annotations) == 0 {
		new.ObjectMeta.Annotations = nil
	}
	patchBytes, err := makeStrategicMergePatch(old, new, v1.Pod{})
	if err != nil {
		return err
	}
	_, err = client.CoreV1().Pods(namespace).Patch(name, types.StrategicMergePatchType, patchBytes)
	// REVISIT: maybe worth a retry in case of version mismatch?
	return err
}

func AddNodeAnnotation(client *kubernetes.Clientset, name, key, value string) error {
	old, err := client.CoreV1().Nodes().Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	new := old.DeepCopy()
	if new.ObjectMeta.Annotations == nil {
		new.ObjectMeta.Annotations = make(map[string]string)
	}
	new.ObjectMeta.Annotations[key] = value
	patchBytes, err := makeStrategicMergePatch(old, new, v1.Node{})
	if err != nil {
		return err
	}
	_, err = client.CoreV1().Nodes().Patch(name, types.StrategicMergePatchType, patchBytes)
	// REVISIT: maybe worth a retry in case of version mismatch?
	return err
}
