package node

import (
	log "github.com/sirupsen/logrus"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/informers"

	"github.com/yamt/midonet-kubernetes/pkg/controller"
	"github.com/yamt/midonet-kubernetes/pkg/midonet"
)

type Handler struct {
}

func NewController(si informers.SharedInformerFactory, kc *kubernetes.Clientset, config *midonet.Config) *controller.Controller {
	informer := si.Core().V1().Nodes().Informer()
	return controller.NewController("Node", informer, &Handler{})
}

func (h *Handler) Update(key string, obj interface{}) error {
	clog := log.WithFields(log.Fields{
		"key": key,
		"obj": obj,
	})
	clog.Info("On Update")
	return nil
}

func (h *Handler) Delete(key string) error {
	clog := log.WithField("key", key)
	clog.Info("On Delete")
	return nil
}
