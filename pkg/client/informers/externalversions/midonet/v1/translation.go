/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	time "time"

	midonet_v1 "github.com/midonet/midonet-kubernetes/pkg/apis/midonet/v1"
	versioned "github.com/midonet/midonet-kubernetes/pkg/client/clientset/versioned"
	internalinterfaces "github.com/midonet/midonet-kubernetes/pkg/client/informers/externalversions/internalinterfaces"
	v1 "github.com/midonet/midonet-kubernetes/pkg/client/listers/midonet/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// TranslationInformer provides access to a shared informer and lister for
// Translations.
type TranslationInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.TranslationLister
}

type translationInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewTranslationInformer constructs a new informer for Translation type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewTranslationInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredTranslationInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredTranslationInformer constructs a new informer for Translation type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredTranslationInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.MidonetV1().Translations(namespace).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.MidonetV1().Translations(namespace).Watch(options)
			},
		},
		&midonet_v1.Translation{},
		resyncPeriod,
		indexers,
	)
}

func (f *translationInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredTranslationInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *translationInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&midonet_v1.Translation{}, f.defaultInformer)
}

func (f *translationInformer) Lister() v1.TranslationLister {
	return v1.NewTranslationLister(f.Informer().GetIndexer())
}
