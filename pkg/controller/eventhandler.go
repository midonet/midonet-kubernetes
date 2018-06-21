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

package controller

import (
	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// AddHandler creates and adds an event handler to the informer.  It returns
// the queue associated to the event handler.
func AddHandler(informer cache.SharedIndexInformer, kind string) workqueue.RateLimitingInterface {
	rateLimiter := workqueue.DefaultControllerRateLimiter()
	queue := workqueue.NewNamedRateLimitingQueue(rateLimiter, kind)
	handler := NewEventHandler(kind, queue)
	informer.AddEventHandler(handler)
	return queue
}

// NewEventHandler creates an event handler which just adds events to
// the given queue.
func NewEventHandler(kind string, queue workqueue.Interface) cache.ResourceEventHandler {
	handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			logAndQueue("Add", kind, queue, obj, nil)
		},
		UpdateFunc: func(old, new interface{}) {
			logAndQueue("Update", kind, queue, new, old)
		},
		DeleteFunc: func(obj interface{}) {
			logAndQueue("Delete", kind, queue, obj, nil)
		},
	}
	return handler
}

func logAndQueue(op string, kind string, queue workqueue.Interface, obj interface{}, oldObj interface{}) error {
	clog := log.WithFields(log.Fields{
		"op":   op,
		"kind": kind,
	})
	// NOTE(yamamoto): For some reasons, client-go uses namespace/name
	// strings for keys, rather than UIDs.  It might cause subtle issues
	// when you delete and create resources with the same name quickly.
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		log.WithError(err).Fatal("DeletionHandlingMetaNamespaceKeyFunc")
	}
	clog = clog.WithFields(log.Fields{
		"key": key,
	})
	if _, ok := obj.(cache.DeletedFinalStateUnknown); !ok {
		meta, err := meta.Accessor(obj)
		if err != nil {
			clog.WithError(err).Fatal("meta.Accessor")
		} else {
			clog = clog.WithFields(log.Fields{
				"uid":     meta.GetUID(),
				"version": meta.GetResourceVersion(),
			})
		}
	}
	if oldObj != nil {
		metaOld, err := meta.Accessor(oldObj)
		if err != nil {
			clog.WithError(err).Fatal("meta.Accessor")
		}
		clog = clog.WithFields(log.Fields{
			"oldVersion": metaOld.GetResourceVersion(),
		})
	}
	clog.Debug("Queueing")
	queue.Add(key)
	return nil
}
