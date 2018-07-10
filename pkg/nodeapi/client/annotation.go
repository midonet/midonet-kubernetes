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

package client

import (
	"errors"
	"net"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/midonet/midonet-kubernetes/pkg/nodeapi"
)

func newClient() (nodeapi.MidoNetKubeNodeClient, error) {
	unixDialer := func(a string, t time.Duration) (net.Conn, error) {
		return net.Dial("unix", a)
	}
	conn, err := grpc.Dial(nodeapi.Path, grpc.WithInsecure(), grpc.WithDialer(unixDialer))
	if err != nil {
		return nil, err
	}
	return nodeapi.NewMidoNetKubeNodeClient(conn), nil
}

func AddPodAnnotation(namespace, name, key, value string) error {
	client, err := newClient()
	if err != nil {
		return err
	}
	req := &nodeapi.AddPodAnnotationRequest{
		Namespace: namespace,
		Name:      name,
		Key:       key,
		Value:     value,
	}
	reply, err := client.AddPodAnnotation(context.Background(), req)
	if err != nil {
		return err
	}
	if reply.Error != "" {
		return errors.New(reply.Error)
	}
	return nil
}

func DeletePodAnnotation(namespace, name, key string) error {
	client, err := newClient()
	if err != nil {
		return err
	}
	req := &nodeapi.DeletePodAnnotationRequest{
		Namespace: namespace,
		Name:      name,
		Key:       key,
	}
	reply, err := client.DeletePodAnnotation(context.Background(), req)
	if err != nil {
		return err
	}
	if reply.Error != "" {
		return errors.New(reply.Error)
	}
	return nil
}
