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

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Handler is a set of callbacks to process events on the queue.
type Handler interface {
	Update(string, schema.GroupVersionKind, interface{}) error
	Delete(string) error
}

// Controller describes a controller to watch the given GVK events.
type Controller struct {
	informer cache.SharedIndexInformer
	queue    workqueue.RateLimitingInterface
	handler  Handler
	gvk      schema.GroupVersionKind
}

// NewController creates a controller.
func NewController(gvk schema.GroupVersionKind, informer cache.SharedIndexInformer, handler Handler) *Controller {
	queue := AddHandler(informer, gvk.String())
	return &Controller{
		informer: informer,
		queue:    queue,
		handler:  handler,
		gvk:      gvk,
	}
}

// Run executes the controller.
func (c *Controller) Run() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	queue := c.queue
	key, quit := queue.Get()
	if quit {
		return false
	}
	defer queue.Done(key)
	clog := log.WithFields(log.Fields{
		"key": key,
	})

	clog.Debug("Start processing.")
	err := c.processItem(key.(string), c.informer)
	if err == nil {
		clog.Debug("Done.")
		queue.Forget(key)
		return true
	}
	clog.WithError(err).Error("Failed. Retrying.")
	queue.AddRateLimited(key)
	return true
}

func (c *Controller) processItem(key string, informer cache.SharedIndexInformer) error {
	clog := log.WithField("key", key)
	clog.Debug("Processing")
	obj, exists, err := informer.GetIndexer().GetByKey(key)
	if err != nil {
		clog.WithError(err).Fatal("GetBykey")
	}
	if !exists {
		clog.Debug("Deleted.")
		return c.handler.Delete(key)
	}
	clog.WithField("obj", obj).Debug("Updated.")
	return c.handler.Update(key, c.gvk, obj)
}

// GetQueue returns the event queue used by the controller.
// It can be useful to inject extra events for the controller.
// See pkg/converter/endpoints/controller.go for an example.
func (c *Controller) GetQueue() workqueue.Interface {
	return c.queue
}
