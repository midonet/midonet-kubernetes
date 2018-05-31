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

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/midonet/midonet-kubernetes/pkg/apis/midonet/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// TranslationLister helps list Translations.
type TranslationLister interface {
	// List lists all Translations in the indexer.
	List(selector labels.Selector) (ret []*v1.Translation, err error)
	// Translations returns an object that can list and get Translations.
	Translations(namespace string) TranslationNamespaceLister
	TranslationListerExpansion
}

// translationLister implements the TranslationLister interface.
type translationLister struct {
	indexer cache.Indexer
}

// NewTranslationLister returns a new TranslationLister.
func NewTranslationLister(indexer cache.Indexer) TranslationLister {
	return &translationLister{indexer: indexer}
}

// List lists all Translations in the indexer.
func (s *translationLister) List(selector labels.Selector) (ret []*v1.Translation, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Translation))
	})
	return ret, err
}

// Translations returns an object that can list and get Translations.
func (s *translationLister) Translations(namespace string) TranslationNamespaceLister {
	return translationNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// TranslationNamespaceLister helps list and get Translations.
type TranslationNamespaceLister interface {
	// List lists all Translations in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1.Translation, err error)
	// Get retrieves the Translation from the indexer for a given namespace and name.
	Get(name string) (*v1.Translation, error)
	TranslationNamespaceListerExpansion
}

// translationNamespaceLister implements the TranslationNamespaceLister
// interface.
type translationNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Translations in the indexer for a given namespace.
func (s translationNamespaceLister) List(selector labels.Selector) (ret []*v1.Translation, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Translation))
	})
	return ret, err
}

// Get retrieves the Translation from the indexer for a given namespace and name.
func (s translationNamespaceLister) Get(name string) (*v1.Translation, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("translation"), name)
	}
	return obj.(*v1.Translation), nil
}
