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

package converter

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"

	"github.com/midonet/midonet-kubernetes/pkg/apis/midonet/v1"
	mncli "github.com/midonet/midonet-kubernetes/pkg/client/clientset/versioned"
)

type TranslationUpdater struct {
	client mncli.Interface
}

func NewTranslationUpdater(client mncli.Interface) *TranslationUpdater {
	return &TranslationUpdater{
		client: client,
	}
}

func (u *TranslationUpdater) Update(parentKind schema.GroupVersionKind, parentObj interface{}, resources map[string][]BackendResource) error {
	var prefix string
	var owners []metav1.OwnerReference
	var ownerlabels map[string]string
	var requirement *labels.Requirement
	var ns string
	hasNamespace := true
	if parentObj != nil {
		prefix = strings.ToLower(parentKind.Kind)
		pmeta, err := meta.Accessor(parentObj)
		if err != nil {
			log.WithError(err).Error("Accessor")
			return err
		}
		puid := pmeta.GetUID()
		owners = []metav1.OwnerReference{
			{
				APIVersion: parentKind.GroupVersion().String(),
				Kind:       parentKind.Kind,
				Name:       pmeta.GetName(),
				UID:        puid,
			},
		}
		ownerlabels = map[string]string{OwnerUIDLabel: string(puid)}
		r, err := labels.NewRequirement(OwnerUIDLabel, selection.Equals, []string{string(puid)})
		if err != nil {
			return err
		}
		requirement = r
		ns = pmeta.GetNamespace()
		if ns == "" {
			// Namespace-less resource like Node.  Use the default
			// namespace for the corresponding Translation resources.
			hasNamespace = false
			ns = metav1.NamespaceDefault
		}
	} else {
		// Note: this prefix should be unique enough so that it won't
		// collide with possible future k8s resource types.
		prefix = "midonet-global"
		owners = nil
		ownerlabels = map[string]string{GlobalLabel: ""}
		r, err := labels.NewRequirement(GlobalLabel, selection.Exists, nil)
		if err != nil {
			return err
		}
		requirement = r
		ns = metav1.NamespaceSystem
	}
	finalizers := []string{MidoNetAPIDeleter}
	var uids []types.UID
	for k, res := range resources {
		// TODO: Stop extracting name from key
		name, err := extractName(k, hasNamespace)
		if err != nil {
			return err
		}
		name = fmt.Sprintf("%s.%s", prefix, name)
		name = makeDNS(name)
		uid, err := u.updateOne(ns, name, owners, ownerlabels, finalizers, res)
		if err != nil {
			return err
		}
		uids = append(uids, uid)
	}
	// Remove stale translations
	return u.deleteTranslations(requirement, uids)
}

func (u *TranslationUpdater) deleteTranslations(req *labels.Requirement, keepUIDs []types.UID) error {
	// Get a list of Translations owned by the parentUID synchronously.
	// REVISIT: Maybe it's more efficient to use the cache in the informer
	// but it might be tricky to avoid races with ourselves:
	//   consider updating a Service twice.
	//   the first update adds a Translation and the second update
	//   deletes it. when the controller processes the second update,
	//   it's possible that its informer have not seen the Translation
	//   addtion from the first update yet. in that case, it might
	//   fail to delete the Translation.
	selector := labels.NewSelector()
	selector = selector.Add(*req)
	opts := metav1.ListOptions{LabelSelector: selector.String()}
	objList, err := u.client.MidonetV1().Translations(metav1.NamespaceAll).List(opts)
	if err != nil {
		return err
	}
	for _, tr := range objList.Items {
		for _, keep := range keepUIDs {
			if tr.ObjectMeta.UID == keep {
				goto next
			}
		}
		err = u.deleteTranslation(tr)
		if err != nil {
			return err
		}
	next:
	}
	return nil
}

func (u *TranslationUpdater) deleteTranslation(tr v1.Translation) error {
	namespace := tr.ObjectMeta.Namespace
	name := tr.ObjectMeta.Name
	err := u.client.MidonetV1().Translations(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"namespace": namespace,
		"name":      name,
	}).Info("Deleted CR")
	return nil
}

func (u *TranslationUpdater) updateOne(ns, name string, owners []metav1.OwnerReference, labels map[string]string, finalizers []string, resources []BackendResource) (types.UID, error) {
	clog := log.WithFields(log.Fields{
		"namespace": ns,
		"name":      name,
	})
	var v1rs []v1.BackendResource
	for _, res := range resources {
		r, err := ToAPI(res)
		if err != nil {
			return "", err
		}
		v1rs = append(v1rs, *r)
	}
	obj := &v1.Translation{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Resources: v1rs,
	}
	meta, err := meta.Accessor(obj)
	if err != nil {
		clog.WithError(err).Error("Accessor(obj)")
		return "", err
	}
	meta.SetOwnerReferences(owners)
	meta.SetLabels(labels)
	meta.SetFinalizers(finalizers)
	clog = clog.WithField("obj", obj)
	newObj, err := u.client.MidonetV1().Translations(ns).Create(obj)
	if err == nil {
		log.WithFields(log.Fields{
			"namespace": ns,
			"name":      name,
		}).Info("Created Translation")
		return newObj.ObjectMeta.UID, nil
	}
	if !errors.IsAlreadyExists(err) {
		clog.WithError(err).Error("Create")
		return "", err
	}
	// NOTE: CRs have AllowUnconditionalUpdate=false
	// REVISIT: Probably should use Patch to avoid overwriting unrelated fields
	existingObj, err := u.client.MidonetV1().Translations(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		clog.WithError(err).Error("Get")
		return "", err
	}
	obj.ObjectMeta.ResourceVersion = existingObj.ObjectMeta.ResourceVersion
	newObj, err = u.client.MidonetV1().Translations(ns).Update(obj)
	if err != nil {
		clog.WithError(err).Error("Update")
		return "", err
	}
	log.WithFields(log.Fields{
		"namespace": ns,
		"name":      name,
	}).Info("Updated Translation")
	return newObj.ObjectMeta.UID, nil
}

func extractName(key string, hasNamespace bool) (string, error) {
	expected := 1
	if hasNamespace {
		expected = 2
	}
	sep := strings.SplitN(key, "/", expected)
	if len(sep) != expected {
		return "", fmt.Errorf("Unrecognized key %s", key)
	}
	return sep[expected-1], nil
}

func makeDNS(name string) string {
	n := strings.Replace(name, "/", "-", -1)
	n = strings.ToLower(n)
	if name != n {
		h := sha1.New()
		h.Write([]byte(name))
		n = fmt.Sprintf("%s-%s", n, hex.EncodeToString(h.Sum(nil)))
	}
	return n
}
