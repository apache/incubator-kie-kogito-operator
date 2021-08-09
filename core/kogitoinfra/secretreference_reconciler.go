// Copyright 2020 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kogitoinfra

import (
	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

type secretReferenceReconciler struct {
	infraContext
	secretHandler infrastructure.SecretHandler
}

func initSecretReferenceReconciler(context infraContext) Reconciler {
	context.Log = context.Log.WithValues("resource", "SecretReference")
	return &secretReferenceReconciler{
		infraContext:  context,
		secretHandler: infrastructure.NewSecretHandler(context.Context),
	}
}

// AppendSecretWatchedObjects ...
func AppendSecretWatchedObjects(b *builder.Builder) *builder.Builder {
	return b.Owns(&corev1.Secret{})
}

// Reconcile reconcile Kogito infra object
func (i *secretReferenceReconciler) Reconcile() error {
	for _, secretReference := range i.instance.GetSpec().GetSecretReferences() {
		if len(secretReference.GetName()) > 0 {
			i.Log.Debug("Custom Secret instance reference is provided")
			namespace := i.instance.GetNamespace()
			secretInstance, resultErr := i.secretHandler.FetchSecret(types.NamespacedName{Name: secretReference.GetName(), Namespace: namespace})
			if resultErr != nil {
				return resultErr
			}
			if secretInstance == nil {
				return errorForResourceNotFound("Secret", secretReference.GetName(), namespace)
			}
		} else {
			return errorForResourceConfigError(i.instance, "No Secret resource name given")
		}
		i.updateSecretReferenceInStatus(secretReference)
	}
	return nil
}

func (i *secretReferenceReconciler) updateSecretReferenceInStatus(secretReference api.SecretReferenceInterface) {
	secretReferences := append(i.instance.GetStatus().GetSecretReferences(), secretReference)
	i.instance.GetStatus().SetSecretReferences(secretReferences)
}
