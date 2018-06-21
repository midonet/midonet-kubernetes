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
	"encoding/json"
	"fmt"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	log "github.com/sirupsen/logrus"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	mnv1 "github.com/midonet/midonet-kubernetes/pkg/apis/midonet/v1"
	mncli "github.com/midonet/midonet-kubernetes/pkg/client/clientset/versioned"
	"github.com/midonet/midonet-kubernetes/pkg/k8s"
)

type translationUpdater struct {
	client   mncli.Interface
	recorder record.EventRecorder
}

// NewTranslationUpdater returns an updater to store Translation resources.
func NewTranslationUpdater(client mncli.Interface, recorder record.EventRecorder) Updater {
	return &translationUpdater{
		client:   client,
		recorder: recorder,
	}
}

func (u *translationUpdater) Update(parentKind schema.GroupVersionKind, parentObjInterface interface{}, resources map[Key][]BackendResource) error {
	var parentObj runtime.Object
	var parentRef *v1.ObjectReference
	// REVISIT: Make the caller pass runtime.Object
	if parentObjInterface != nil {
		parentObj = parentObjInterface.(runtime.Object)
		ref, err := k8s.GetReferenceForEvent(parentObj)
		if err != nil {
			return err
		}
		parentRef = ref
	} else {
		parentObj = nil
		parentRef = nil
	}
	var owners []metav1.OwnerReference
	var ownerlabels map[string]string
	var requirement *labels.Requirement
	var ns string
	if parentObj != nil {
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
			ns = metav1.NamespaceDefault
		}
	} else {
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
		name := k.translationName()
		name = makeDNS(name)
		uid, err := u.updateOne(parentRef, ns, name, owners, ownerlabels, finalizers, res)
		if err != nil {
			return err
		}
		uids = append(uids, uid)
	}
	// Remove stale translations
	return u.deleteTranslations(parentRef, requirement, uids)
}

func (u *translationUpdater) deleteTranslations(parentRef *v1.ObjectReference, req *labels.Requirement, keepUIDs []types.UID) error {
	// Get a list of Translations owned by the parentUID synchronously.
	// REVISIT: Maybe it's more efficient to use the cache in the informer
	// but it might be tricky to avoid races with ourselves:
	//   consider updating a Service twice.
	//   the first update adds a Translation and the second update
	//   deletes it. when the controller processes the second update,
	//   it's possible that its informer have not seen the Translation
	//   addition from the first update yet. in that case, it might
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
		if parentRef != nil {
			u.recorder.Eventf(parentRef, v1.EventTypeNormal, "TranslationDeleted", "Translation %s/%s Deleted", tr.ObjectMeta.Namespace, tr.ObjectMeta.Name)
		} else {
			log.WithFields(log.Fields{
				"namespace": tr.ObjectMeta.Namespace,
				"name":      tr.ObjectMeta.Name,
			}).Info("Global Translation Deleted")
		}
	next:
	}
	return nil
}

func (u *translationUpdater) deleteTranslation(tr mnv1.Translation) error {
	namespace := tr.ObjectMeta.Namespace
	name := tr.ObjectMeta.Name
	err := u.client.MidonetV1().Translations(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (u *translationUpdater) updateOne(parentRef *v1.ObjectReference, ns, name string, owners []metav1.OwnerReference, labels map[string]string, finalizers []string, resources []BackendResource) (types.UID, error) {
	clog := log.WithFields(log.Fields{
		"namespace": ns,
		"name":      name,
	})
	var v1rs []mnv1.BackendResource
	for _, res := range resources {
		r, err := toAPI(res)
		if err != nil {
			return "", err
		}
		v1rs = append(v1rs, *r)
	}
	obj := &mnv1.Translation{
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
		if parentRef != nil {
			u.recorder.Eventf(parentRef, v1.EventTypeNormal, "TranslationCreated", "Translation %s/%s Created", ns, name)
		} else {
			log.WithFields(log.Fields{
				"namespace": ns,
				"name":      name,
			}).Info("Global Translation Created")
		}
		return newObj.ObjectMeta.UID, nil
	}
	if !errors.IsAlreadyExists(err) {
		clog.WithError(err).Error("Create")
		return "", err
	}
	// NOTE: CRs have AllowUnconditionalUpdate=false
	// NOTE: CRs don't support strategic merge patch
	existingObj, err := u.client.MidonetV1().Translations(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		clog.WithError(err).Error("Get")
		return "", err
	}
	oldData, err := json.Marshal(existingObj)
	if err != nil {
		return "", err
	}
	desiredObj := existingObj.DeepCopy()
	desiredObj.Resources = obj.Resources
	desiredData, err := json.Marshal(desiredObj)
	if err != nil {
		return "", err
	}
	patchBytes, err := jsonpatch.CreateMergePatch(oldData, desiredData)
	if err != nil {
		return "", err
	}
	clog = clog.WithField("patch", string(patchBytes))
	newObj, err = u.client.MidonetV1().Translations(ns).Patch(name, types.MergePatchType, patchBytes)
	if err != nil {
		clog.WithError(err).Error("Patch")
		return "", err
	}
	if jsonpatch.Equal(patchBytes, []byte(`{}`)) {
		log.WithFields(log.Fields{
			"namespace": ns,
			"name":      name,
		}).Debug("Skipping no-op update of Translation")
		return existingObj.ObjectMeta.UID, nil
	}
	checkTranslationUpdate(existingObj, desiredObj)
	log.WithFields(log.Fields{
		"namespace": ns,
		"name":      name,
		"patch":     string(patchBytes),
	}).Debug("Patched Translation")
	if parentRef != nil {
		u.recorder.Eventf(parentRef, v1.EventTypeNormal, "TranslationUpdated", "Translation %s/%s Updated", ns, name)
	} else {
		log.WithFields(log.Fields{
			"namespace": ns,
			"name":      name,
		}).Info("Global Translation Updated")
	}
	return newObj.ObjectMeta.UID, nil
}

func checkTranslationUpdate(old *mnv1.Translation, new *mnv1.Translation) {
	clog := log.WithFields(log.Fields{
		"old": "old",
		"new": "new",
	})
	if len(old.Resources) > len(new.Resources) {
		clog.Fatal("The list of resources shrunk")
	}
}

// makeDNS tweaks the given name so that it's usable as a Kubernetes
// resource name.
// REVISIT: probabaly we should truncate it when too long.
func makeDNS(name string) string {
	n := strings.Replace(name, "/", "-", -1)
	n = strings.ToLower(n)
	// If we changed anything, append the hash of the original string
	// to maintain uniqueness.
	if name != n {
		h := sha1.New()
		h.Write([]byte(name))
		n = fmt.Sprintf("%s-%s", n, hex.EncodeToString(h.Sum(nil)))
	}
	return n
}
