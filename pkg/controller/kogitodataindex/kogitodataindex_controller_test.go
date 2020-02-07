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

package kogitodataindex

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitodataindex/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	routev1 "github.com/openshift/api/route/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/stretchr/testify/assert"

	utilsres "github.com/RHsyseng/operator-utils/pkg/resource"
)

func TestReconcileKogitoDataIndex_Reconcile(t *testing.T) {
	ns := t.Name()
	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: ns,
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			InfinispanMeta: v1alpha1.InfinispanMeta{InfinispanProperties: v1alpha1.InfinispanConnectionProperties{UseKogitoInfra: true}},
		},
	}
	kafkaList := &kafkabetav1.KafkaList{
		Items: []kafkabetav1.Kafka{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kafka",
					Namespace: ns,
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
	r := &ReconcileKogitoDataIndex{
		client: cli,
		scheme: meta.GetRegisteredSchema(),
	}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}

	// basic checks
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	// check infra
	infra, ready, err := infrastructure.EnsureKogitoInfra(ns, cli).WithInfinispan().Apply()
	assert.NoError(t, err)
	assert.False(t, ready)  // we don't have status defined since the KogitoInfra controller is not running
	assert.NotNil(t, infra) // we have a infra instance created during reconciliation phase
	assert.Equal(t, infrastructure.DefaultKogitoInfraName, infra.GetName())
}

func TestReconcileKogitoDataIndex_UpdateHTTPPort(t *testing.T) {
	ns := t.Name()
	envMap := make(map[string]string)
	envMap[resource.DataIndexEnvKeyHTTPPort] = "3030"
	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: ns,
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			HTTPPort: 9090,
			Env:      envMap,
			KafkaMeta: v1alpha1.KafkaMeta{
				KafkaProperties: v1alpha1.KafkaConnectionProperties{
					UseKogitoInfra: false,
					ExternalURI:    "my-uri:9022",
				},
			},
			InfinispanMeta: v1alpha1.InfinispanMeta{
				InfinispanProperties: v1alpha1.InfinispanConnectionProperties{
					UseKogitoInfra: false,
					URI:            "another-uri:11222",
				},
			},
		},
	}

	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance}, nil, nil)
	r := &ReconcileKogitoDataIndex{
		client: cli,
		scheme: meta.GetRegisteredSchema(),
	}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}

	// first reconcile
	_, err := r.Reconcile(req)
	assert.NoError(t, err)

	// make sure HTTPPort env was added on the deployment
	deployment := &appsv1.Deployment{}
	exists, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, deployment)
	assert.True(t, exists)
	assert.NoError(t, err)

	// make sure that the http port was correctly added.
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  resource.DataIndexEnvKeyHTTPPort,
		Value: "9090",
	})

	assert.Equal(t, int32(9090), deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort)
	assert.Equal(t, int32(9090), deployment.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket.Port.IntVal)
	assert.Equal(t, int32(9090), deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.TCPSocket.Port.IntVal)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: instance.Name, Namespace: instance.Namespace},
	}
	exists, err = kubernetes.ResourceC(cli).Fetch(service)
	assert.True(t, exists)
	assert.NoError(t, err)
	assert.Equal(t, int32(9090), service.Spec.Ports[0].TargetPort.IntVal)

	// update the route
	// reconcile and test
	// compare the route http port
	routeFromResource := &routev1.Route{}
	routeFound, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, routeFromResource)
	assert.NoError(t, err)
	assert.True(t, routeFound)
	// update http port on the given route
	routeFromResource.Spec.Port.TargetPort.IntVal = 4000
	err = kubernetes.ResourceC(cli).Update(routeFromResource)
	assert.NoError(t, err)
	// reconcile
	_, err = r.Reconcile(req)
	assert.NoError(t, err)
	// get the route after reconcile
	routeAfterReconcile := &routev1.Route{}
	routeAfterReconcileFound, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, routeAfterReconcile)
	assert.True(t, routeAfterReconcileFound)
	assert.NoError(t, err)
	//fmt.Println(routeAfterReconcile)
	assert.True(t, routeAfterReconcileFound)
	assert.Equal(t, intstr.IntOrString{Type: 0, IntVal: 9090, StrVal: ""}, routeAfterReconcile.Spec.Port.TargetPort)

	// update the service
	// reconcile and test
	// compare the service http and target port
	serviceFromResource := &corev1.Service{}
	serviceFound, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, serviceFromResource)
	assert.True(t, serviceFound)
	assert.NoError(t, err)
	//fmt.Println(serviceFromResource)
	// update ports
	serviceFromResource.Spec.Ports[0].Port = 4000
	serviceFromResource.Spec.Ports[0].TargetPort = intstr.FromString("4000")
	err = kubernetes.ResourceC(cli).Update(serviceFromResource)
	assert.NoError(t, err)
	// reconcile
	_, err = r.Reconcile(req)
	assert.NoError(t, err)
	// get the service after reconcile
	serviceAfterReconcile := &corev1.Service{}
	servicefterReconcileFound, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, serviceAfterReconcile)
	assert.True(t, servicefterReconcileFound)
	assert.NoError(t, err)
	//fmt.Println(serviceAfterReconcile)
	// compare again if the port was updated after reconcile
	assert.Equal(t, int32(9090), serviceAfterReconcile.Spec.Ports[0].Port)
	assert.Equal(t, intstr.FromInt(9090), serviceAfterReconcile.Spec.Ports[0].TargetPort)
}

func Test_getKubernetesResources(t *testing.T) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "ss1",
		},
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "svc1",
		},
	}

	rt := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "rt1",
		},
	}

	kt1 := &kafkabetav1.KafkaTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "kt1",
		},
	}

	kt2 := &kafkabetav1.KafkaTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "kt2",
		},
	}

	type args struct {
		resources *resource.KogitoDataIndexResources
	}
	tests := []struct {
		name string
		args args
		want []utilsres.KubernetesResource
	}{
		{
			"GetKubernetesResources",
			args{
				&resource.KogitoDataIndexResources{
					Deployment:  deployment,
					Service:     svc,
					Route:       rt,
					KafkaTopics: []*kafkabetav1.KafkaTopic{kt1, kt2},
				},
			},
			[]utilsres.KubernetesResource{deployment, svc, rt, kt1, kt2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getKubernetesResources(tt.args.resources); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getKubernetesResources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReconcileKogitoDataIndex_updateStatus(t *testing.T) {
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

	resourcesNoUpdate := false
	var noErr error
	resourcesUpdate := true
	err := fmt.Errorf("test error")

	type fields struct {
		client *client.Client
		scheme *runtime.Scheme
	}
	type args struct {
		request         *reconcile.Request
		instance        *appv1alpha1.KogitoDataIndex
		resources       *resource.KogitoDataIndexResources
		resourcesUpdate *bool
		err             *error
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *v1alpha1.KogitoDataIndexStatus
		wantErr error
	}{
		{
			"updateStatus",
			fields{
				test.CreateFakeClient([]runtime.Object{deployment, svc, rt,
					&v1alpha1.KogitoDataIndex{
						Spec: v1alpha1.KogitoDataIndexSpec{
							HTTPPort: 8080,
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
				meta.GetRegisteredSchema(),
			},
			args{
				&reconcile.Request{},
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
				&resourcesNoUpdate,
				&noErr,
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
			nil,
		},
		{
			"resourceUpdate",
			fields{
				test.CreateFakeClient([]runtime.Object{deployment, svc, rt,
					&v1alpha1.KogitoDataIndex{
						Spec: v1alpha1.KogitoDataIndexSpec{
							HTTPPort: 9090,
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
				meta.GetRegisteredSchema(),
			},
			args{
				&reconcile.Request{},
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
				&resourcesUpdate,
				&noErr,
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
			nil,
		},
		{
			"reconcileError",
			fields{
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
				meta.GetRegisteredSchema(),
			},
			args{
				&reconcile.Request{},
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
				&resourcesNoUpdate,
				&err,
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
			err,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReconcileKogitoDataIndex{
				client: tt.fields.client,
				scheme: tt.fields.scheme,
			}
			r.updateStatus(tt.args.request, tt.args.instance, tt.args.resources, tt.args.resourcesUpdate, tt.args.err)

			if !reflect.DeepEqual(*tt.args.err, tt.wantErr) {
				t.Errorf("ManageStatus() error = %v, wantErr %v", *tt.args.err, tt.wantErr)
				return
			}

			if exist, err := kubernetes.ResourceC(tt.fields.client).Fetch(tt.args.instance); err != nil {
				t.Errorf("updateStatus() failed to update data index instance status, error = %v", err)
			} else if !exist {
				t.Errorf("updateStatus() failed to retrieve data index instance")
			} else {
				if !reflect.DeepEqual(tt.args.instance.Status.DeploymentStatus, tt.want.DeploymentStatus) ||
					!reflect.DeepEqual(tt.args.instance.Status.ServiceStatus, tt.want.ServiceStatus) ||
					!reflect.DeepEqual(tt.args.instance.Status.Route, tt.want.Route) ||
					!reflect.DeepEqual(tt.args.instance.Status.DependenciesStatus, tt.want.DependenciesStatus) ||
					len(tt.args.instance.Status.Conditions) != len(tt.want.Conditions) ||
					!reflect.DeepEqual(tt.args.instance.Status.Conditions[0].Condition, tt.want.Conditions[0].Condition) ||
					!reflect.DeepEqual(tt.args.instance.Status.Conditions[0].Message, tt.want.Conditions[0].Message) {
					t.Errorf("updateStatus() got status = %v, want status %v", tt.args.instance.Status, tt.want)
				}
			}
		})
	}
}
