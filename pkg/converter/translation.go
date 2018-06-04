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

func (u *TranslationUpdater) Update(key string, parentKind schema.GroupVersionKind, parentObj interface{}, resources map[string][]BackendResource) error {
	var uids []types.UID
	prefix := strings.ToLower(parentKind.Kind)
	pmeta, err := meta.Accessor(parentObj)
	if err != nil {
		log.WithError(err).Error("Accessor")
		return err
	}
	puid := pmeta.GetUID()
	owners := []metav1.OwnerReference{
		{
			APIVersion: parentKind.GroupVersion().String(),
			Kind:       parentKind.Kind,
			Name:       pmeta.GetName(),
			UID:        puid,
		},
	}
	labels := map[string]string{OwnerUIDLabel: string(puid)}
	finalizers := []string{MidoNetAPIDeleter}
	for k, res := range resources {
		uid, err := u.updateOne(k, prefix, owners, labels, finalizers, res)
		if err != nil {
			return err
		}
		uids = append(uids, uid)
	}
	// Remove stale translations
	return u.deleteTranslations(puid, uids)
}

func (u *TranslationUpdater) deleteTranslations(parentUID types.UID, keepUIDs []types.UID) error {
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
	req, err := labels.NewRequirement(OwnerUIDLabel, selection.Equals, []string{string(parentUID)})
	if err != nil {
		return err
	}
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

func (u *TranslationUpdater) updateOne(key string, prefix string, owners []metav1.OwnerReference, labels map[string]string, finalizers []string, resources []BackendResource) (types.UID, error) {
	ns, name, err := extractNames(key)
	if err != nil {
		return "", err
	}
	name = fmt.Sprintf("%s.%s", prefix, name)
	name = makeDNS(name)
	clog := log.WithFields(log.Fields{
		"key":       key,
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

func (u *TranslationUpdater) Delete(key string) error {
	/* nothing to do */
	return nil
}

func extractNames(key string) (string, string, error) {
	sep := strings.SplitN(key, "/", 2)
	if len(sep) == 1 {
		// Probably a namespace-less resource like Node.
		// Use the default namespace.
		return metav1.NamespaceDefault, sep[0], nil
	}
	if len(sep) != 2 {
		return "", "", fmt.Errorf("Unrecognized key %s", key)
	}
	return sep[0], sep[1], nil
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
