package record

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_generateEvent(t *testing.T) {
	service := &v1alpha1.KogitoJobsService{
		ObjectMeta: v1.ObjectMeta{
			Name:      "jobs-service",
			Namespace: t.Name(),
		},
	}
	cli := test.CreateFakeClientOnOpenShift(nil, nil, nil)
	recorder := NewRecorder(meta.GetRegisteredSchema(), corev1.EventSource{Component: service.GetName()})
	recorder.Eventf(cli, service, "Normal", "Created", "Create Deployment")

	eventList := &corev1.EventList{}
	kubernetes.ResourceC(cli).ListWithNamespace(t.Name(), eventList)
	assert.NotNil(t, eventList.Items)
	assert.Equal(t, 1, len(eventList.Items))
}

func Test_generateEvent_InvalidEventType(t *testing.T) {
	service := &v1alpha1.KogitoJobsService{
		ObjectMeta: v1.ObjectMeta{
			Name:      "jobs-service",
			Namespace: t.Name(),
		},
	}
	cli := test.CreateFakeClientOnOpenShift(nil, nil, nil)
	recorder := NewRecorder(meta.GetRegisteredSchema(), corev1.EventSource{Component: service.GetName()})
	recorder.Eventf(cli, service, "InvalidEventType", "Created", "Create Deployment")

	eventList := &corev1.EventList{}
	kubernetes.ResourceC(cli).ListWithNamespace(t.Name(), eventList)
	assert.Nil(t, eventList.Items)
}
