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

package framework

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/test/config"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sv1beta1 "k8s.io/api/extensions/v1beta1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	kubeClient *client.Client
	mux        = &sync.Mutex{}
)

// podErrorReasons contains all the reasons to state a pod in error.
var podErrorReasons = [1]string{"ErrImagePull"}

// InitKubeClient initializes the Kubernetes Client
func InitKubeClient() error {
	mux.Lock()
	defer mux.Unlock()
	if kubeClient == nil {
		newClient, err := client.NewClientBuilder().UseControllerDynamicMapper().WithDiscoveryClient().WithBuildClient().WithKubernetesExtensionClient().Build()
		if err != nil {
			return fmt.Errorf("Error initializing kube client: %v", err)
		}
		kubeClient = newClient
	}
	return nil
}

// WaitForPodsWithLabel waits for pods with specific label to be available and running
func WaitForPodsWithLabel(namespace, labelName, labelValue string, numberOfPods, timeoutInMin int) error {
	return WaitForOnOpenshift(namespace, fmt.Sprintf("Pods with label name '%s' and value '%s' available and running", labelName, labelValue), timeoutInMin,
		func() (bool, error) {
			pods, err := GetPodsWithLabels(namespace, map[string]string{labelName: labelValue})
			if err != nil || (len(pods.Items) != numberOfPods) {
				return false, err
			}

			return CheckPodsAreReady(pods), nil
		}, CheckPodsWithLabelInError(namespace, labelName, labelValue))
}

// GetPods retrieves all pods in namespace
func GetPods(namespace string) (*corev1.PodList, error) {
	pods := &corev1.PodList{}
	if err := kubernetes.ResourceC(kubeClient).ListWithNamespace(namespace, pods); err != nil {
		return nil, err
	}
	return pods, nil
}

// GetPodsByDeploymentConfig retrieves pods with a deploymentconfig label set to <dcName>
func GetPodsByDeploymentConfig(namespace string, dcName string) (*corev1.PodList, error) {
	return GetPodsWithLabels(namespace, map[string]string{"deploymentconfig": dcName})
}

// GetPodsByDeployment retrieves pods belonging to a Deployment
func GetPodsByDeployment(namespace string, dName string) (pods []corev1.Pod, err error) {
	pods = []corev1.Pod{}

	// Get ReplicaSet related to the Deployment
	replicaSet, err := GetActiveReplicaSetByDeployment(namespace, dName)
	if err != nil {
		return nil, err
	}

	// Fetch all pods in namespace
	podList := &corev1.PodList{}
	if err := kubernetes.ResourceC(kubeClient).ListWithNamespace(namespace, podList); err != nil {
		return nil, err
	}

	// Find which pods belong to the ReplicaSet
	for _, pod := range podList.Items {
		for _, ownerReference := range pod.OwnerReferences {
			if ownerReference.Kind == "ReplicaSet" && ownerReference.Name == replicaSet.GetName() {
				pods = append(pods, pod)
			}
		}
	}

	return
}

// GetActiveReplicaSetByDeployment retrieves active ReplicaSet belonging to a Deployment
func GetActiveReplicaSetByDeployment(namespace string, dName string) (*apps.ReplicaSet, error) {
	replicaSets := &apps.ReplicaSetList{}
	if err := kubernetes.ResourceC(kubeClient).ListWithNamespace(namespace, replicaSets); err != nil {
		return nil, err
	}

	// Find ReplicaSet owned by Deployment with active Pods
	for _, replicaSet := range replicaSets.Items {
		for _, ownerReference := range replicaSet.OwnerReferences {
			if ownerReference.Kind == "Deployment" && ownerReference.Name == dName && replicaSet.Spec.Size() > 0 {
				return &replicaSet, nil
			}
		}
	}

	return nil, fmt.Errorf("No ReplicaSet belonging to Deployment %s found", dName)
}

// GetPodsWithLabels retrieves pods based on label name and value
func GetPodsWithLabels(namespace string, labels map[string]string) (*corev1.PodList, error) {
	pods := &corev1.PodList{}
	if err := kubernetes.ResourceC(kubeClient).ListWithNamespaceAndLabel(namespace, pods, labels); err != nil {
		return nil, err
	}
	return pods, nil
}

// CheckPodsAreReady returns true if all pods are ready
func CheckPodsAreReady(pods *corev1.PodList) bool {
	for _, pod := range pods.Items {
		if !IsPodStatusConditionReady(&pod) {
			return false
		}
	}
	return true
}

// IsPodRunning returns true if pod is running
func IsPodRunning(pod *corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodRunning
}

// IsPodStatusConditionReady returns true if all pod's containers are ready (really running)
func IsPodStatusConditionReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.ContainersReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

// WaitForDeploymentRunning waits for a deployment to be running, with a specific number of pod
func WaitForDeploymentRunning(namespace, dName string, podNb int, timeoutInMin int) error {
	return WaitForOnOpenshift(namespace, fmt.Sprintf("Deployment %s running", dName), timeoutInMin,
		func() (bool, error) {
			if dc, err := GetDeployment(namespace, dName); err != nil {
				return false, err
			} else if dc == nil {
				return false, nil
			} else {
				GetLogger(namespace).Debug("Deployment has", "available replicas", dc.Status.AvailableReplicas)
				return dc.Status.Replicas == int32(podNb) && dc.Status.AvailableReplicas == int32(podNb), nil
			}
		})
}

// GetDeployment retrieves deployment with specified name in namespace
func GetDeployment(namespace, deploymentName string) (*apps.Deployment, error) {
	deployment := &apps.Deployment{}
	if exists, err := kubernetes.ResourceC(kubeClient).FetchWithKey(types.NamespacedName{Name: deploymentName, Namespace: namespace}, deployment); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}
	return deployment, nil
}

func loadResource(namespace, uri string, resourceRef meta.ResourceObject, beforeCreate func(object interface{})) error {
	GetLogger(namespace).Debug("loadResource", "uri", uri)

	data, err := ReadFromURI(uri)
	if err != nil {
		return fmt.Errorf("Unable to read from URI %s: %v", uri, err)
	}

	if err = kubernetes.ResourceC(kubeClient).CreateFromYamlContent(data, namespace, resourceRef, beforeCreate); err != nil {
		return fmt.Errorf("Error while creating resources from file '%s': %v ", uri, err)
	}
	return nil
}

// WaitForAllPodsByDeploymentConfigToContainTextInLog waits for pods of specified deployment config to contain specified text in log
func WaitForAllPodsByDeploymentConfigToContainTextInLog(namespace, dcName, logText string, timeoutInMin int) error {
	return WaitForOnOpenshift(namespace, fmt.Sprintf("Pods for deployment config '%s' contain text '%s'", dcName, logText), timeoutInMin,
		func() (bool, error) {
			pods, err := GetPodsByDeploymentConfig(namespace, dcName)
			if err != nil {
				return false, err
			}

			// Container name is equal to deployment config name
			return checkAllPodsContainingTextInLog(namespace, pods.Items, dcName, logText)
		}, CheckPodsByDeploymentConfigInError(namespace, dcName))
}

// WaitForAllPodsByDeploymentToContainTextInLog waits for pods of specified deployment to contain specified text in log
func WaitForAllPodsByDeploymentToContainTextInLog(namespace, dName, logText string, timeoutInMin int) error {
	return waitForPodsByDeploymentToContainTextInLog(namespace, dName, logText, timeoutInMin, checkAllPodsContainingTextInLog)
}

// WaitForAnyPodsByDeploymentToContainTextInLog waits for pods of specified deployment to contain specified text in log
func WaitForAnyPodsByDeploymentToContainTextInLog(namespace, dName, logText string, timeoutInMin int) error {
	return waitForPodsByDeploymentToContainTextInLog(namespace, dName, logText, timeoutInMin, checkAnyPodsContainingTextInLog)
}

func waitForPodsByDeploymentToContainTextInLog(namespace, dName, logText string, timeoutInMin int, predicate func(string, []corev1.Pod, string, string) (bool, error)) error {
	return WaitForOnOpenshift(namespace, fmt.Sprintf("Pods for deployment '%s' contain text '%s'", dName, logText), timeoutInMin,
		func() (bool, error) {
			pods, err := GetPodsByDeployment(namespace, dName)
			if err != nil {
				return false, err
			}

			// Container name is equal to deployment config name
			return predicate(namespace, pods, dName, logText)
		}, CheckPodsByDeploymentInError(namespace, dName))
}

func checkAnyPodsContainingTextInLog(namespace string, pods []corev1.Pod, containerName, text string) (bool, error) {
	for _, pod := range pods {
		containsText, err := isPodContainingTextInLog(namespace, &pod, containerName, text)
		if err != nil {
			return false, err
		} else if containsText {
			return true, nil
		}
	}

	return false, nil
}

func checkAllPodsContainingTextInLog(namespace string, pods []corev1.Pod, containerName, text string) (bool, error) {
	for _, pod := range pods {
		containsText, err := isPodContainingTextInLog(namespace, &pod, containerName, text)
		if err != nil || !containsText {
			return false, err
		}
	}
	return true, nil
}

func isPodContainingTextInLog(namespace string, pod *corev1.Pod, containerName, text string) (bool, error) {
	log, err := kubernetes.PodC(kubeClient).GetLogs(namespace, pod.GetName(), containerName)
	return strings.Contains(log, text), err
}

// IsCrdAvailable returns whether the crd is available on cluster
func IsCrdAvailable(crdName string) (bool, error) {
	crdEntity := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: crdName,
		},
	}
	return kubernetes.ResourceC(kubeClient).Fetch(crdEntity)
}

// DeleteObject deletes object
func DeleteObject(o meta.ResourceObject) error {
	return kubernetes.ResourceC(kubeClient).Delete(o)
}

// CreateSecret creates a new secret
func CreateSecret(namespace, name string, secretContent map[string]string) error {
	GetLogger(namespace).Info("Create Secret %s", "name", name)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: secretContent,
	}

	return kubernetes.ResourceC(kubeClient).Create(secret)
}

// CheckPodHasImagePullSecretWithPrefix checks that a pod has an image pull secret starting with the given prefix
func CheckPodHasImagePullSecretWithPrefix(pod *corev1.Pod, imagePullSecretPrefix string) bool {
	for _, secretRef := range pod.Spec.ImagePullSecrets {
		if strings.HasPrefix(secretRef.Name, imagePullSecretPrefix) {
			return true
		}
	}
	return false
}

// CheckPodsByDeploymentConfigInError returns a function that checks the pods error state.
func CheckPodsByDeploymentConfigInError(namespace string, dcName string) func() (bool, error) {
	return func() (bool, error) {
		pods, err := GetPodsByDeploymentConfig(namespace, dcName)
		if err != nil {
			return true, err

		}
		return checkPodsInError(pods.Items)
	}
}

// CheckPodsByDeploymentInError returns a function that checks the pods error state.
func CheckPodsByDeploymentInError(namespace string, dName string) func() (bool, error) {
	return func() (bool, error) {
		pods, err := GetPodsByDeployment(namespace, dName)
		if err != nil {
			return true, err

		}
		return checkPodsInError(pods)
	}
}

// CheckPodsWithLabelInError returns a function that checks the pods error state.
func CheckPodsWithLabelInError(namespace, labelName, labelValue string) func() (bool, error) {
	return func() (bool, error) {
		pods, err := GetPodsWithLabels(namespace, map[string]string{labelName: labelValue})
		if err != nil {
			return true, err

		}
		return checkPodsInError(pods.Items)
	}
}

func checkPodsInError(pods []corev1.Pod) (bool, error) {
	for _, pod := range pods {
		if hasErrors, err := isPodInError(&pod); hasErrors {
			return true, err
		}
	}

	return false, nil
}

func isPodInError(pod *corev1.Pod) (bool, error) {
	if IsPodRunning(pod) {
		return false, nil
	}

	for _, status := range pod.Status.ContainerStatuses {
		for _, reason := range podErrorReasons {
			if status.State.Waiting != nil && status.State.Waiting.Reason == reason {
				return true, fmt.Errorf("Error in pod, reason: %s", reason)
			}
		}

	}

	return false, nil
}

func checkPodContainerHasEnvVariableWithValue(pod *corev1.Pod, containerName, envVarName, envVarValue string) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			for _, env := range container.Env {
				if env.Name == envVarName {
					return env.Value == envVarValue
				}
			}
		}
	}
	return false
}

// GetIngressURI returns the ingress URI
func GetIngressURI(namespace, serviceName string) (string, error) {
	ingress := &k8sv1beta1.Ingress{}
	if exists, err := kubernetes.ResourceC(kubeClient).FetchWithKey(types.NamespacedName{Name: serviceName, Namespace: namespace}, ingress); err != nil {
		return "", err
	} else if !exists {
		return "", fmt.Errorf("Ingress %s does not exist in namespace %s", serviceName, namespace)
	} else if len(ingress.Spec.Rules) == 0 {
		return "", fmt.Errorf("Ingress %s does not have any rules", serviceName)
	}

	return fmt.Sprintf("http://%s:80", ingress.Spec.Rules[0].Host), nil
}

// ExposeServiceOnKubernetes adds ingress CR to expose a service
func ExposeServiceOnKubernetes(namespace string, service v1beta1.KogitoService) error {
	host := service.GetName()
	if !config.IsLocalCluster() {
		host += fmt.Sprintf(".%s.%s", namespace, config.GetDomainSuffix())
	}

	port := framework.DefaultExposedPort

	ingress := k8sv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        service.GetName(),
			Namespace:   namespace,
			Annotations: map[string]string{"nginx.ingress.kubernetes.io/rewrite-target": "/"},
		},
		Spec: k8sv1beta1.IngressSpec{
			Rules: []k8sv1beta1.IngressRule{
				{
					Host: host,
					IngressRuleValue: k8sv1beta1.IngressRuleValue{
						HTTP: &k8sv1beta1.HTTPIngressRuleValue{
							Paths: []k8sv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: k8sv1beta1.IngressBackend{
										ServiceName: service.GetName(),
										ServicePort: intstr.FromInt(port),
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return kubernetes.ResourceC(kubeClient).Create(&ingress)
}

// WaitForOnKubernetes is a specific method
func WaitForOnKubernetes(namespace, display string, timeoutInMin int, condition func() (bool, error)) error {
	return WaitFor(namespace, display, GetKubernetesDurationFromTimeInMin(timeoutInMin), condition)
}

// GetKubernetesDurationFromTimeInMin will calculate the time depending on the configured cluster load factor
func GetKubernetesDurationFromTimeInMin(timeoutInMin int) time.Duration {
	return time.Duration(timeoutInMin*config.GetLoadFactor()) * time.Minute
}

// IsOpenshift returns whether the cluster is running on Openshift
func IsOpenshift() bool {
	return kubeClient.IsOpenshift()
}
