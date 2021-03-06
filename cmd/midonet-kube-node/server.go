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

package main

import (
	"net"
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"

	"github.com/midonet/midonet-kubernetes/pkg/converter"
	"github.com/midonet/midonet-kubernetes/pkg/k8s"
	api "github.com/midonet/midonet-kubernetes/pkg/nodeapi"
)

type server struct {
	client *kubernetes.Clientset
}

func (s *server) AddPodAnnotation(ctx context.Context, in *api.AddPodAnnotationRequest) (*api.AddPodAnnotationReply, error) {
	logger := log.WithFields(log.Fields{
		"request": "AddPodAnnotation",
		"args":    in,
	})

	logger.Info("Got a request")
	var errorMessage string
	var reason string
	if in.Key == converter.MACAnnotation {
		err := k8s.AddPodAnnotation(s.client, in.Namespace, in.Name, in.Key, in.Value)
		if err != nil {
			errorMessage = err.Error()
			reason = string(errors.ReasonForError(err))
			logger.WithError(err).WithField("reason", reason).Error("Failed")
		} else {
			logger.Info("Succeed")
		}
	} else {
		logger.Error("Rejected")
		errorMessage = "Rejected"
	}
	return &api.AddPodAnnotationReply{
		Error:              errorMessage,
		Metav1StatusReason: reason,
	}, nil
}

func (s *server) DeletePodAnnotation(ctx context.Context, in *api.DeletePodAnnotationRequest) (*api.DeletePodAnnotationReply, error) {
	logger := log.WithFields(log.Fields{
		"request": "DeletePodAnnotation",
		"args":    in,
	})

	logger.Info("Got a request")
	var errorMessage string
	var reason string
	if in.Key == converter.MACAnnotation {
		err := k8s.DeletePodAnnotation(s.client, in.Namespace, in.Name, in.Key)
		if err != nil {
			errorMessage = err.Error()
			reason = string(errors.ReasonForError(err))
			logger.WithError(err).WithField("reason", reason).Error("Failed")
		} else {
			logger.Info("Succeed")
		}
	} else {
		logger.Error("Rejected")
		errorMessage = "Rejected"
	}
	return &api.DeletePodAnnotationReply{
		Error:              errorMessage,
		Metav1StatusReason: reason,
	}, nil
}

func serveRPC(clientset *kubernetes.Clientset) {
	log.Info("Starting RPC server")
	logger := log.WithField("path", api.Path)
	os.Remove(api.Path)
	logger.Info("Listening")
	l, err := net.Listen("unix", api.Path)
	if err != nil {
		logger.WithError(err).Fatal("Failed to listen")
	}
	s := grpc.NewServer()
	api.RegisterMidoNetKubeNodeServer(s, &server{client: clientset})
	logger.Info("Serving")
	err = s.Serve(l)
	if err != nil {
		logger.WithError(err).Fatal("Failed to serve")
	}
	logger.Fatal("RPC server exited")
}
