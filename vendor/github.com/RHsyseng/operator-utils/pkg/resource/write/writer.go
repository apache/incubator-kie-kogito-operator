package write

import (
	"context"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/write/hooks"
	newerror "github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type UpdateHooks interface {
	Trigger(existing resource.KubernetesResource, requested resource.KubernetesResource) error
}

type resourceWriter struct {
	writer          clientv1.Writer
	ownerRefs       []metav1.OwnerReference
	ownerController metav1.Object
	scheme          *runtime.Scheme
	updateHooks     UpdateHooks
}

// New creates a resourceWriter object that can be used to add/update/remove kubernetes resources
// the provided writer object will be used for the underlying operations
func New(writer clientv1.Writer) *resourceWriter {
	return &resourceWriter{
		writer:      writer,
		updateHooks: hooks.DefaultUpdateHooks(),
	}
}

// WithOwnerReferences allows owner references to be set on any object that's added or updated
// calling this function removes any owner controller that may have been configured
func (this *resourceWriter) WithOwnerReferences(ownerRefs ...metav1.OwnerReference) *resourceWriter {
	this.ownerRefs = ownerRefs
	this.ownerController = nil
	this.scheme = nil
	return this
}

// WithOwnerController allows a controlling owner to be set on any object that's added or updated
// calling this function removes the effect of previously setting owner references on this resource writer
func (this *resourceWriter) WithOwnerController(ownerController metav1.Object, scheme *runtime.Scheme) *resourceWriter {
	this.ownerController = ownerController
	this.scheme = scheme
	this.ownerRefs = nil
	return this
}

// WithCustomUpdateHooks allows intercepting update calls to set missing fields
// default provided update hooks, for example, set the resource version and GVK based on existing counterpart
func (this *resourceWriter) WithCustomUpdateHooks(updateHooks UpdateHooks) *resourceWriter {
	this.updateHooks = updateHooks
	return this
}

// AddResources sets ownership as/if configured, and then uses the writer to create them
// the boolean result is true if any changes were made
func (this *resourceWriter) AddResources(resources []resource.KubernetesResource) (bool, error) {
	var added bool
	for index := range resources {
		requested := resources[index]
		if this.ownerRefs != nil {
			requested.SetOwnerReferences(this.ownerRefs)
		} else if this.ownerController != nil {
			err := controllerutil.SetControllerReference(this.ownerController, requested, this.scheme)
			if err != nil {
				return added, err
			}
		}
		err := this.writer.Create(context.TODO(), requested)
		if err != nil {
			return added, err
		}
		added = true
	}
	return added, nil
}

// UpdateResources finds the updated counterpart for each of the provided resources in the existing array and uses it to set resource version and GVK
// It also sets ownership as/if configured, and then uses the writer to update them
// the boolean result is true if any changes were made
func (this *resourceWriter) UpdateResources(existing []resource.KubernetesResource, resources []resource.KubernetesResource) (bool, error) {
	var updated bool
	for index := range resources {
		requested := resources[index]
		var counterpart resource.KubernetesResource
		for _, candidate := range existing {
			if candidate.GetNamespace() == requested.GetNamespace() && candidate.GetName() == requested.GetName() {
				counterpart = candidate
				break
			}
		}
		if counterpart == nil {
			return updated, newerror.New("Failed to find a deployed counterpart to resource being updated")
		}
		err := this.updateHooks.Trigger(counterpart, requested)
		if err != nil {
			return updated, err
		}
		if this.ownerRefs != nil {
			requested.SetOwnerReferences(this.ownerRefs)
		} else if this.ownerController != nil {
			err := controllerutil.SetControllerReference(this.ownerController, requested, this.scheme)
			if err != nil {
				return updated, err
			}
		}
		err = this.writer.Update(context.TODO(), requested)
		if err != nil {
			return updated, err
		}
		updated = true
	}
	return updated, nil
}

// RemoveResources removes each of the provided resources using the provided writer
// the boolean result is true if any changes were made
func (this *resourceWriter) RemoveResources(resources []resource.KubernetesResource) (bool, error) {
	var removed bool
	for index := range resources {
		err := this.writer.Delete(context.TODO(), resources[index])
		if err != nil {
			return removed, err
		}
		removed = true
	}
	return removed, nil
}
