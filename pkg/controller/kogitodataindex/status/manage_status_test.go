// Copyright 2019 Red Hat, Inc. and/or its affiliates
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

package status

import (
	"reflect"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitodataindex/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"

	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	routev1 "github.com/openshift/api/route/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_ManageStatus_WhenTheresStatusChange(t *testing.T) {
	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			Replicas: 1,
			KafkaMeta: v1alpha1.KafkaMeta{
				KafkaProperties: v1alpha1.KafkaConnectionProperties{
					Instance: "kafka",
				},
			},
		},
	}
	kafkaList := &kafkabetav1.KafkaList{
		Items: []kafkabetav1.Kafka{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kafka",
					Namespace: "test",
				},
				Spec: kafkabetav1.KafkaSpec{
					Kafka: kafkabetav1.KafkaClusterSpec{
						Replicas: 1,
					},
				},
				Status: kafkabetav1.KafkaStatus{
					Listeners: []kafkabetav1.ListenerStatus{
						{
							Type: "plain",
							Addresses: []kafkabetav1.ListenerAddress{
								{
									Host: "kafka",
									Port: 9092,
								},
							},
						},
					},
				},
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{instance, kafkaList}, nil, nil)
	resources, err := resource.GetRequestedResources(instance, cli)
	assert.NoError(t, err)

	err = ManageStatus(instance, resources, false, false, cli)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.NotNil(t, instance.Status.Conditions)
	assert.Len(t, instance.Status.Conditions, 1)
}

func Test_checkCurrentCondition(t *testing.T) {
	type args struct {
		deployment      *appsv1.Deployment
		service         *corev1.Service
		resourcesUpdate bool
		reconcileError  bool
	}
	tests := []struct {
		name string
		args args
		want v1alpha1.DataIndexCondition
	}{
		{
			"ReconcileError",
			args{
				&appsv1.Deployment{},
				&corev1.Service{},
				false,
				true,
			},
			v1alpha1.DataIndexCondition{
				Condition: v1alpha1.ConditionFailed,
				Message:   "Deployment Error",
			},
		},
		{
			"NoDeployment",
			args{
				nil,
				&corev1.Service{},
				false,
				false,
			},
			v1alpha1.DataIndexCondition{
				Condition: v1alpha1.ConditionFailed,
				Message:   "Deployment Failed",
			},
		},
		{
			"NoService",
			args{
				&appsv1.Deployment{},
				nil,
				false,
				false,
			},
			v1alpha1.DataIndexCondition{
				Condition: v1alpha1.ConditionFailed,
				Message:   "Deployment Failed",
			},
		},
		{
			"ResourcesUpdate",
			args{
				&appsv1.Deployment{},
				&corev1.Service{},
				true,
				false,
			},
			v1alpha1.DataIndexCondition{
				Condition: v1alpha1.ConditionProvisioning,
				Message:   "Deploying Objects",
			},
		},
		{
			"Provisioning",
			args{
				&appsv1.Deployment{
					Status: appsv1.DeploymentStatus{
						ReadyReplicas: 0,
						Replicas:      1,
					},
				},
				&corev1.Service{},
				false,
				false,
			},
			v1alpha1.DataIndexCondition{
				Condition: v1alpha1.ConditionProvisioning,
				Message:   "Deployment In Progress",
			},
		},
		{
			"OK",
			args{
				&appsv1.Deployment{
					Status: appsv1.DeploymentStatus{
						ReadyReplicas: 1,
						Replicas:      1,
					},
				},
				&corev1.Service{},
				false,
				false,
			},
			v1alpha1.DataIndexCondition{
				Condition: v1alpha1.ConditionOK,
				Message:   "Deployment Finished",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkCurrentCondition(tt.args.deployment, tt.args.service, tt.args.resourcesUpdate, tt.args.reconcileError); !reflect.DeepEqual(got.Condition, tt.want.Condition) || !reflect.DeepEqual(got.Message, tt.want.Message) {
				t.Errorf("checkCurrentCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManageStatus(t *testing.T) {
	deploymentStatus := appsv1.DeploymentStatus{
		Replicas:      1,
		ReadyReplicas: 1,
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "ss1",
		},
		Status: deploymentStatus,
	}

	svcStatus := corev1.ServiceStatus{
		LoadBalancer: corev1.LoadBalancerStatus{
			Ingress: []corev1.LoadBalancerIngress{
				{
					Hostname: "test",
					IP:       "1.1.1.1",
				},
			},
		},
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "svc1",
		},
		Status: svcStatus,
	}

	rt := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "rt1",
		},
		Spec: routev1.RouteSpec{
			Host: "test",
		},
	}

	type args struct {
		instance        *v1alpha1.KogitoDataIndex
		resources       *resource.KogitoDataIndexResources
		resourcesUpdate bool
		reconcileError  bool
		client          *client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    *v1alpha1.KogitoDataIndexStatus
		wantErr bool
	}{
		{
			"ManageStatus",
			args{
				&v1alpha1.KogitoDataIndex{
					Spec: v1alpha1.KogitoDataIndexSpec{
						InfinispanMeta: v1alpha1.InfinispanMeta{
							InfinispanProperties: v1alpha1.InfinispanConnectionProperties{
								URI: "test-infinispan",
							},
						},
						KafkaMeta: v1alpha1.KafkaMeta{
							KafkaProperties: v1alpha1.KafkaConnectionProperties{
								ExternalURI: "test-kafka",
							},
						},
					},
				},
				&resource.KogitoDataIndexResources{
					Deployment: deployment,
					Service:    svc,
					Route:      rt,
				},
				false,
				false,
				test.CreateFakeClient([]runtime.Object{deployment, svc, rt,
					&v1alpha1.KogitoDataIndex{
						Spec: v1alpha1.KogitoDataIndexSpec{
							InfinispanMeta: v1alpha1.InfinispanMeta{
								InfinispanProperties: v1alpha1.InfinispanConnectionProperties{
									URI: "test-infinispan",
								},
							},
							KafkaMeta: v1alpha1.KafkaMeta{
								KafkaProperties: v1alpha1.KafkaConnectionProperties{
									ExternalURI: "test-kafka",
								},
							},
						},
					}},
					nil, nil),
			},
			&v1alpha1.KogitoDataIndexStatus{
				DeploymentStatus: deploymentStatus,
				ServiceStatus:    svcStatus,
				Route:            "http://test",
				DependenciesStatus: []v1alpha1.DataIndexDependenciesStatus{
					v1alpha1.DataIndexDependenciesStatusOK,
				},
				Conditions: []v1alpha1.DataIndexCondition{
					{
						Condition: v1alpha1.ConditionOK,
						Message:   "Deployment Finished",
					},
				},
			},
			false,
		},
		{
			"ResourceUpdate",
			args{
				&v1alpha1.KogitoDataIndex{
					Spec: v1alpha1.KogitoDataIndexSpec{
						InfinispanMeta: v1alpha1.InfinispanMeta{
							InfinispanProperties: v1alpha1.InfinispanConnectionProperties{
								URI: "test-infinispan",
							},
						},
						KafkaMeta: v1alpha1.KafkaMeta{
							KafkaProperties: v1alpha1.KafkaConnectionProperties{
								ExternalURI: "test-kafka",
							},
						},
					},
				},
				&resource.KogitoDataIndexResources{
					Deployment: deployment,
					Service:    svc,
					Route:      rt,
				},
				true,
				false,
				test.CreateFakeClient([]runtime.Object{deployment, svc, rt,
					&v1alpha1.KogitoDataIndex{
						Spec: v1alpha1.KogitoDataIndexSpec{
							InfinispanMeta: v1alpha1.InfinispanMeta{
								InfinispanProperties: v1alpha1.InfinispanConnectionProperties{
									URI: "test-infinispan",
								},
							},
							KafkaMeta: v1alpha1.KafkaMeta{
								KafkaProperties: v1alpha1.KafkaConnectionProperties{
									ExternalURI: "test-kafka",
								},
							},
						},
					}},
					nil, nil),
			},
			&v1alpha1.KogitoDataIndexStatus{
				DeploymentStatus: deploymentStatus,
				ServiceStatus:    svcStatus,
				Route:            "http://test",
				DependenciesStatus: []v1alpha1.DataIndexDependenciesStatus{
					v1alpha1.DataIndexDependenciesStatusOK,
				},
				Conditions: []v1alpha1.DataIndexCondition{
					{
						Condition: v1alpha1.ConditionProvisioning,
						Message:   "Deploying Objects",
					},
				},
			},
			false,
		},
		{
			"ReconcileError",
			args{
				&v1alpha1.KogitoDataIndex{
					Spec: v1alpha1.KogitoDataIndexSpec{
						InfinispanMeta: v1alpha1.InfinispanMeta{
							InfinispanProperties: v1alpha1.InfinispanConnectionProperties{
								URI: "test-infinispan",
							},
						},
						KafkaMeta: v1alpha1.KafkaMeta{
							KafkaProperties: v1alpha1.KafkaConnectionProperties{
								ExternalURI: "test-kafka",
							},
						},
					},
				},
				&resource.KogitoDataIndexResources{
					Deployment: deployment,
					Service:    svc,
					Route:      rt,
				},
				false,
				true,
				test.CreateFakeClient([]runtime.Object{deployment, svc, rt,
					&v1alpha1.KogitoDataIndex{
						Spec: v1alpha1.KogitoDataIndexSpec{
							InfinispanMeta: v1alpha1.InfinispanMeta{
								InfinispanProperties: v1alpha1.InfinispanConnectionProperties{
									URI: "test-infinispan",
								},
							},
							KafkaMeta: v1alpha1.KafkaMeta{
								KafkaProperties: v1alpha1.KafkaConnectionProperties{
									ExternalURI: "test-kafka",
								},
							},
						},
					}},
					nil, nil),
			},
			&v1alpha1.KogitoDataIndexStatus{
				DeploymentStatus: deploymentStatus,
				ServiceStatus:    svcStatus,
				Route:            "http://test",
				DependenciesStatus: []v1alpha1.DataIndexDependenciesStatus{
					v1alpha1.DataIndexDependenciesStatusOK,
				},
				Conditions: []v1alpha1.DataIndexCondition{
					{
						Condition: v1alpha1.ConditionFailed,
						Message:   "Deployment Error",
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ManageStatus(tt.args.instance, tt.args.resources, tt.args.resourcesUpdate, tt.args.reconcileError, tt.args.client); (err != nil) != tt.wantErr {
				t.Errorf("ManageStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})

		if exist, err := kubernetes.ResourceC(tt.args.client).Fetch(tt.args.instance); err != nil {
			t.Errorf("ManageStatus() failed to update data index instance status, error = %v", err)
		} else if !exist {
			t.Errorf("ManageStatus() failed to retrieve data index instance")
		} else {
			if !reflect.DeepEqual(tt.args.instance.Status.DeploymentStatus, tt.want.DeploymentStatus) ||
				!reflect.DeepEqual(tt.args.instance.Status.ServiceStatus, tt.want.ServiceStatus) ||
				!reflect.DeepEqual(tt.args.instance.Status.Route, tt.want.Route) ||
				!reflect.DeepEqual(tt.args.instance.Status.DependenciesStatus, tt.want.DependenciesStatus) ||
				len(tt.args.instance.Status.Conditions) != len(tt.want.Conditions) ||
				!reflect.DeepEqual(tt.args.instance.Status.Conditions[0].Condition, tt.want.Conditions[0].Condition) ||
				!reflect.DeepEqual(tt.args.instance.Status.Conditions[0].Message, tt.want.Conditions[0].Message) {
				t.Errorf("ManageStatus() got status = %v, want status %v", tt.args.instance.Status, tt.want)
			}
		}
	}
}
