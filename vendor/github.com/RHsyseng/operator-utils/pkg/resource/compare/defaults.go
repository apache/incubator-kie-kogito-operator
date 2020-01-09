package compare

import (
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"strings"
)

const (
	imageTriggersAnnotation           = "image.openshift.io/triggers"
	deploymentRevisionAnnotation      = "deployment.kubernetes.io/revision"
	imageTriggerContainerNameValueFmt = "spec.template.spec.containers[?(@.name==\\\"%s\\\")].image"
)

type resourceComparator struct {
	defaultCompareFunc func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool
	compareFuncMap     map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool
}

func (this *resourceComparator) SetDefaultComparator(compFunc func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool) {
	this.defaultCompareFunc = compFunc
}

func (this *resourceComparator) GetDefaultComparator() func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	return this.defaultCompareFunc
}

func (this *resourceComparator) SetComparator(resourceType reflect.Type, compFunc func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool) {
	this.compareFuncMap[resourceType] = compFunc
}

func (this *resourceComparator) GetComparator(resourceType reflect.Type) func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	return this.compareFuncMap[resourceType]
}

func (this *resourceComparator) Compare(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	compareFunc := this.GetDefaultComparator()
	type1 := reflect.ValueOf(deployed).Elem().Type()
	type2 := reflect.ValueOf(requested).Elem().Type()
	if type1 == type2 {
		if comparator, exists := this.compareFuncMap[type1]; exists {
			compareFunc = comparator
		}
	}
	return compareFunc(deployed, requested)
}

func (this *resourceComparator) CompareArrays(deployed []resource.KubernetesResource, requested []resource.KubernetesResource) ResourceDelta {
	deployedMap := getObjectMap(deployed)
	requestedMap := getObjectMap(requested)
	var added []resource.KubernetesResource
	var updated []resource.KubernetesResource
	var removed []resource.KubernetesResource
	for name, requestedObject := range requestedMap {
		deployedObject := deployedMap[name]
		if deployedObject == nil {
			added = append(added, requestedObject)
		} else if !this.Compare(deployedObject, requestedObject) {
			updated = append(updated, requestedObject)
		}
	}
	for name, deployedObject := range deployedMap {
		if requestedMap[name] == nil {
			removed = append(removed, deployedObject)
		}
	}
	return ResourceDelta{
		Added:   added,
		Updated: updated,
		Removed: removed,
	}
}

func getObjectMap(objects []resource.KubernetesResource) map[string]resource.KubernetesResource {
	objectMap := make(map[string]resource.KubernetesResource)
	for index := range objects {
		objectMap[objects[index].GetName()] = objects[index]
	}
	return objectMap
}

func defaultMap() map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	equalsMap := make(map[reflect.Type]func(resource.KubernetesResource, resource.KubernetesResource) bool)
	equalsMap[reflect.TypeOf(oappsv1.DeploymentConfig{})] = equalDeploymentConfigs
	equalsMap[reflect.TypeOf(appsv1.Deployment{})] = equalDeployment
	equalsMap[reflect.TypeOf(corev1.Service{})] = equalServices
	equalsMap[reflect.TypeOf(routev1.Route{})] = equalRoutes
	equalsMap[reflect.TypeOf(rbacv1.Role{})] = equalRoles
	equalsMap[reflect.TypeOf(rbacv1.RoleBinding{})] = equalRoleBindings
	equalsMap[reflect.TypeOf(corev1.ServiceAccount{})] = equalServiceAccounts
	equalsMap[reflect.TypeOf(corev1.Secret{})] = equalSecrets
	equalsMap[reflect.TypeOf(buildv1.BuildConfig{})] = equalBuildConfigs
	return equalsMap
}

func equalDeploymentConfigs(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	dc1 := deployed.(*oappsv1.DeploymentConfig)
	dc2 := requested.(*oappsv1.DeploymentConfig)

	//Removed generated fields from deployed version, when not specified in requested item
	dc1 = dc1.DeepCopy()
	triggerBasedImage := make(map[string]bool)
	if dc2.Spec.Strategy.RecreateParams == nil {
		dc1.Spec.Strategy.RecreateParams = nil
	}
	if dc2.Spec.Strategy.ActiveDeadlineSeconds == nil {
		dc1.Spec.Strategy.ActiveDeadlineSeconds = nil
	}
	if dc2.Spec.Strategy.RollingParams == nil && dc2.Spec.Strategy.Type == "" && dc1.Spec.Strategy.Type == oappsv1.DeploymentStrategyTypeRolling {
		//This looks like a default generated strategy that should be ignored
		dc1.Spec.Strategy.Type = ""
		dc1.Spec.Strategy.RollingParams = nil
	}
	if dc1.Spec.Strategy.RollingParams != nil && dc2.Spec.Strategy.RollingParams != nil {
		if dc2.Spec.Strategy.RollingParams.UpdatePeriodSeconds == nil {
			dc1.Spec.Strategy.RollingParams.UpdatePeriodSeconds = nil
		}
		if dc2.Spec.Strategy.RollingParams.IntervalSeconds == nil {
			dc1.Spec.Strategy.RollingParams.IntervalSeconds = nil
		}
		if dc2.Spec.Strategy.RollingParams.TimeoutSeconds == nil {
			dc1.Spec.Strategy.RollingParams.TimeoutSeconds = nil
		}
		if dc2.Spec.Strategy.RollingParams.MaxUnavailable == nil {
			dc1.Spec.Strategy.RollingParams.MaxUnavailable = nil
		}
		if dc2.Spec.Strategy.RollingParams.MaxSurge == nil {
			dc1.Spec.Strategy.RollingParams.MaxSurge = nil
		}
	}
	if dc2.Spec.RevisionHistoryLimit == nil {
		dc1.Spec.RevisionHistoryLimit = nil
	}
	if len(dc1.Spec.Triggers) == 1 && len(dc2.Spec.Triggers) == 0 {
		defaultTrigger := oappsv1.DeploymentTriggerPolicy{Type: oappsv1.DeploymentTriggerOnConfigChange}
		if dc1.Spec.Triggers[0] == defaultTrigger {
			//Remove default generated trigger
			dc1.Spec.Triggers = nil
		}
	}
	for i := range dc1.Spec.Triggers {
		if len(dc2.Spec.Triggers) <= i {
			logger.Info("No matching trigger found in requested DC", "deployed.DC.trigger", dc1.Spec.Triggers[i])
			return false
		}
		if dc1.Spec.Triggers[i].ImageChangeParams != nil && dc2.Spec.Triggers[i].ImageChangeParams != nil {
			if dc2.Spec.Triggers[i].ImageChangeParams.LastTriggeredImage == "" {
				dc1.Spec.Triggers[i].ImageChangeParams.LastTriggeredImage = ""
			}
			for _, containerName := range dc2.Spec.Triggers[i].ImageChangeParams.ContainerNames {
				triggerBasedImage[containerName] = true
			}
		}
	}
	if !checkGeneratePodValues(dc1.Spec.Template, dc2.Spec.Template, triggerBasedImage) {
		return false
	}
	ignoreEmptyMaps(dc1, dc2)

	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{dc1.Name, dc2.Name})
	pairs = append(pairs, [2]interface{}{dc1.Namespace, dc2.Namespace})
	pairs = append(pairs, [2]interface{}{dc1.Labels, dc2.Labels})
	pairs = append(pairs, [2]interface{}{dc1.Annotations, dc2.Annotations})
	pairs = append(pairs, [2]interface{}{dc1.Spec, dc2.Spec})
	equal := EqualPairs(pairs)
	if !equal {
		logger.Info("Resources are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}

func equalDeployment(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	d1 := deployed.(*appsv1.Deployment)
	d2 := requested.(*appsv1.Deployment)

	d1 = d1.DeepCopy()
	triggerBasedImage := make(map[string]bool)

	if d2.Spec.Strategy.RollingUpdate == nil && d1.Spec.Strategy.RollingUpdate != nil {
		d1.Spec.Strategy.RollingUpdate = nil
	}
	if d1.Spec.Strategy.RollingUpdate != nil && d2.Spec.Strategy.RollingUpdate != nil {
		if d2.Spec.Strategy.RollingUpdate.MaxSurge == nil {
			d1.Spec.Strategy.RollingUpdate.MaxSurge = nil
		}
		if d2.Spec.Strategy.RollingUpdate.MaxUnavailable == nil {
			d1.Spec.Strategy.RollingUpdate.MaxUnavailable = nil
		}
	}
	if d2.Spec.RevisionHistoryLimit == nil {
		d1.Spec.RevisionHistoryLimit = nil
	}
	if d2.Spec.ProgressDeadlineSeconds == nil {
		d1.Spec.ProgressDeadlineSeconds = nil
	}

	if &d1.Spec.Template != nil && &d2.Spec.Template != nil {
		if v, ok := d1.Annotations[imageTriggersAnnotation]; ok {
			for _, container := range d1.Spec.Template.Spec.Containers {
				if strings.Contains(v, fmt.Sprintf(imageTriggerContainerNameValueFmt, container.Name)) {
					triggerBasedImage[container.Name] = true
				}
			}
			for _, initContainer := range d1.Spec.Template.Spec.InitContainers {
				if strings.Contains(v, fmt.Sprintf(imageTriggerContainerNameValueFmt, initContainer.Name)) {
					triggerBasedImage[initContainer.Name] = true
				}
			}
		}
		if !checkGeneratePodValues(&d1.Spec.Template, &d2.Spec.Template, triggerBasedImage) {
			return false
		}
	}

	delete(d1.Annotations, deploymentRevisionAnnotation)

	ignoreEmptyMaps(d1, d2)

	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{d1.Name, d2.Name})
	pairs = append(pairs, [2]interface{}{d1.Namespace, d2.Namespace})
	pairs = append(pairs, [2]interface{}{d1.Labels, d2.Labels})
	pairs = append(pairs, [2]interface{}{d1.Annotations, d2.Annotations})
	pairs = append(pairs, [2]interface{}{d1.Spec, d2.Spec})
	equal := EqualPairs(pairs)
	if !equal {
		logger.Info("Resources are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}

func checkGeneratePodValues(pod1 *corev1.PodTemplateSpec, pod2 *corev1.PodTemplateSpec, triggerBasedImage map[string]bool) bool {
	if pod1 != nil && pod2 != nil {
		for i := range pod1.Spec.Volumes {
			if len(pod2.Spec.Volumes) <= i {
				logger.Info("No matching volume found in requested deployment", "deployed.deployment.volume", pod1.Spec.Volumes[i])
				return false
			}
			volSrc1 := pod1.Spec.Volumes[i].VolumeSource
			volSrc2 := pod2.Spec.Volumes[i].VolumeSource
			if volSrc1.Secret != nil && volSrc2.Secret != nil && volSrc2.Secret.DefaultMode == nil {
				volSrc1.Secret.DefaultMode = nil
			}
		}
		if pod2.Spec.RestartPolicy == "" {
			pod1.Spec.RestartPolicy = ""
		}
		if pod2.Spec.DNSPolicy == "" {
			pod1.Spec.DNSPolicy = ""
		}
		//noinspection GoDeprecation
		if pod2.Spec.DeprecatedServiceAccount == "" {
			pod1.Spec.DeprecatedServiceAccount = ""
		}
		if pod2.Spec.SecurityContext == nil {
			pod1.Spec.SecurityContext = nil
		}
		if pod2.Spec.SchedulerName == "" {
			pod1.Spec.SchedulerName = ""
		}
		ignoreGenerateContainerValues(pod1.Spec.Containers, pod2.Spec.Containers, triggerBasedImage)
		ignoreGenerateContainerValues(pod1.Spec.InitContainers, pod2.Spec.InitContainers, triggerBasedImage)
		if pod2.Spec.TerminationGracePeriodSeconds == nil {
			pod1.Spec.TerminationGracePeriodSeconds = nil
		}
	}
	return true
}

func ignoreGenerateContainerValues(containers1 []corev1.Container, containers2 []corev1.Container, triggerBasedImage map[string]bool) {
	for i := range containers1 {
		if len(containers2) <= i {
			logger.Info("No matching container found in requested resource", "deployed container", containers1[i])
			return
		}
		if containers2[i].ImagePullPolicy == "" && containers1[i].ImagePullPolicy == corev1.PullAlways {
			containers1[i].ImagePullPolicy = ""
		}
		probes1 := []*corev1.Probe{containers1[i].LivenessProbe, containers1[i].ReadinessProbe}
		probes2 := []*corev1.Probe{containers2[i].LivenessProbe, containers2[i].ReadinessProbe}
		for index := range probes1 {
			probe1 := probes1[index]
			probe2 := probes2[index]
			if probe1 != nil && probe2 != nil {
				if probe2.FailureThreshold == 0 {
					probe1.FailureThreshold = probe2.FailureThreshold
				}
				if probe2.SuccessThreshold == 0 {
					probe1.SuccessThreshold = probe2.SuccessThreshold
				}
				if probe2.PeriodSeconds == 0 {
					probe1.PeriodSeconds = probe2.PeriodSeconds
				}
				if probe2.TimeoutSeconds == 0 {
					probe1.TimeoutSeconds = probe2.TimeoutSeconds
				}
			}
		}
		if containers2[i].TerminationMessagePath == "" {
			containers1[i].TerminationMessagePath = ""
		}
		if containers2[i].TerminationMessagePolicy == "" {
			containers1[i].TerminationMessagePolicy = ""
		}
		for j := range containers1[i].Env {
			if len(containers2[i].Env) <= j {
				return
			}
			valueFrom := containers2[i].Env[j].ValueFrom
			if valueFrom != nil && valueFrom.FieldRef != nil && valueFrom.FieldRef.APIVersion == "" {
				valueFrom1 := containers1[i].Env[j].ValueFrom
				if valueFrom1 != nil && valueFrom1.FieldRef != nil {
					valueFrom1.FieldRef.APIVersion = ""
				}
			}
		}
		if triggerBasedImage[containers1[i].Name] {
			//Image is being derived from the ImageChange trigger, so this image field is auto-generated after deployment
			containers1[i].Image = containers2[i].Image
		}
	}

}

func equalServices(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	service1 := deployed.(*corev1.Service)
	service2 := requested.(*corev1.Service)

	//Removed generated fields from deployed version, when not specified in requested item
	service1 = service1.DeepCopy()

	//Removed potentially generated annotations for cert request
	delete(service1.Annotations, "service.alpha.openshift.io/serving-cert-signed-by")
	delete(service1.Annotations, "service.beta.openshift.io/serving-cert-signed-by")
	if service2.Spec.ClusterIP == "" {
		service1.Spec.ClusterIP = ""
	}
	if service2.Spec.Type == "" {
		service1.Spec.Type = ""
	}
	if service2.Spec.SessionAffinity == "" {
		service1.Spec.SessionAffinity = ""
	}
	for _, port2 := range service2.Spec.Ports {
		if found, port1 := findServicePort(port2, service1.Spec.Ports); found {
			if port2.Protocol == "" {
				port1.Protocol = ""
			}
		}
	}
	ignoreEmptyMaps(service1, service2)

	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{service1.Name, service2.Name})
	pairs = append(pairs, [2]interface{}{service1.Namespace, service2.Namespace})
	pairs = append(pairs, [2]interface{}{service1.Labels, service2.Labels})
	pairs = append(pairs, [2]interface{}{service1.Annotations, service2.Annotations})
	pairs = append(pairs, [2]interface{}{service1.Spec, service2.Spec})
	equal := EqualPairs(pairs)
	if !equal {
		logger.Info("Resources are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}

func findServicePort(port corev1.ServicePort, ports []corev1.ServicePort) (bool, *corev1.ServicePort) {
	for index, candidate := range ports {
		if port.Name == candidate.Name {
			return true, &ports[index]
		}
	}
	return false, &corev1.ServicePort{}
}

func equalRoutes(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	route1 := deployed.(*routev1.Route)
	route2 := requested.(*routev1.Route)
	route1 = route1.DeepCopy()

	//Removed generated fields from deployed version, that are not specified in requested item
	delete(route1.GetAnnotations(), "openshift.io/host.generated")
	if route2.Spec.Host == "" {
		route1.Spec.Host = ""
	}
	if route2.Spec.To.Kind == "" {
		route1.Spec.To.Kind = ""
	}
	if route2.Spec.To.Name == "" {
		route1.Spec.To.Name = ""
	}
	if route2.Spec.To.Weight == nil {
		route1.Spec.To.Weight = nil
	}
	if route2.Spec.WildcardPolicy == "" {
		route1.Spec.WildcardPolicy = ""
	}
	ignoreEmptyMaps(route1, route2)

	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{route1.Name, route2.Name})
	pairs = append(pairs, [2]interface{}{route1.Namespace, route2.Namespace})
	pairs = append(pairs, [2]interface{}{route1.Labels, route2.Labels})
	pairs = append(pairs, [2]interface{}{route1.Annotations, route2.Annotations})
	pairs = append(pairs, [2]interface{}{route1.Spec, route2.Spec})
	equal := EqualPairs(pairs)
	if !equal {
		logger.Info("Resources are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}

func equalRoles(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	role1 := deployed.(*rbacv1.Role)
	role2 := requested.(*rbacv1.Role)
	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{role1.Name, role2.Name})
	pairs = append(pairs, [2]interface{}{role1.Namespace, role2.Namespace})
	pairs = append(pairs, [2]interface{}{role1.Labels, role2.Labels})
	pairs = append(pairs, [2]interface{}{role1.Annotations, role2.Annotations})
	pairs = append(pairs, [2]interface{}{role1.Rules, role2.Rules})
	equal := EqualPairs(pairs)
	if !equal {
		logger.Info("Resources are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}

func equalServiceAccounts(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	sa1 := deployed.(*corev1.ServiceAccount)
	sa2 := requested.(*corev1.ServiceAccount)
	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{sa1.Name, sa2.Name})
	pairs = append(pairs, [2]interface{}{sa1.Namespace, sa2.Namespace})
	pairs = append(pairs, [2]interface{}{sa1.Labels, sa2.Labels})
	pairs = append(pairs, [2]interface{}{sa1.Annotations, sa2.Annotations})
	equal := EqualPairs(pairs)
	if !equal {
		logger.Info("Resources are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}

func equalRoleBindings(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	binding1 := deployed.(*rbacv1.RoleBinding)
	binding2 := requested.(*rbacv1.RoleBinding)
	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{binding1.Name, binding2.Name})
	pairs = append(pairs, [2]interface{}{binding1.Namespace, binding2.Namespace})
	pairs = append(pairs, [2]interface{}{binding1.Labels, binding2.Labels})
	pairs = append(pairs, [2]interface{}{binding1.Annotations, binding2.Annotations})
	pairs = append(pairs, [2]interface{}{binding1.Subjects, binding2.Subjects})
	pairs = append(pairs, [2]interface{}{binding1.RoleRef.Name, binding2.RoleRef.Name})
	equal := EqualPairs(pairs)
	if !equal {
		logger.Info("Resources are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}

func equalSecrets(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	secret1 := deployed.(*corev1.Secret)
	secret2 := requested.(*corev1.Secret)
	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{secret1.Name, secret2.Name})
	pairs = append(pairs, [2]interface{}{secret1.Namespace, secret2.Namespace})
	pairs = append(pairs, [2]interface{}{secret1.Labels, secret2.Labels})
	pairs = append(pairs, [2]interface{}{secret1.Annotations, secret2.Annotations})
	pairs = append(pairs, [2]interface{}{secret1.Data, secret2.Data})
	pairs = append(pairs, [2]interface{}{secret1.StringData, secret2.StringData})
	equal := EqualPairs(pairs)
	if !equal {
		logger.Info("Resources are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}

func equalBuildConfigs(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	bc1 := deployed.(*buildv1.BuildConfig)
	bc2 := requested.(*buildv1.BuildConfig)

	//Removed generated fields from deployed version, when not specified in requested item
	bc1 = bc1.DeepCopy()
	bc2 = bc2.DeepCopy()
	if bc2.Spec.RunPolicy == "" {
		bc1.Spec.RunPolicy = ""
	}
	for i := range bc1.Spec.Triggers {
		if len(bc1.Spec.Triggers) <= i {
			return false
		}
		trigger1 := bc1.Spec.Triggers[i]
		trigger2 := bc2.Spec.Triggers[i]
		webHooks1 := []*buildv1.WebHookTrigger{trigger1.BitbucketWebHook, trigger1.GenericWebHook, trigger1.GitHubWebHook, trigger1.GitLabWebHook}
		webHooks2 := []*buildv1.WebHookTrigger{trigger2.BitbucketWebHook, trigger2.GenericWebHook, trigger2.GitHubWebHook, trigger2.GitLabWebHook}
		for index := range webHooks1 {
			if webHooks1[index] != nil && webHooks2[index] != nil {
				//noinspection GoDeprecation
				if webHooks1[index].Secret != "" && webHooks2[index].Secret != "" {
					webHooks1[index].Secret = ""
					webHooks2[index].Secret = ""
				}
				if webHooks1[index].SecretReference != nil && webHooks2[index].SecretReference != nil {
					webHooks1[index].SecretReference = nil
					webHooks2[index].SecretReference = nil
				}
			}
		}
		if trigger2.ImageChange != nil && trigger1.ImageChange != nil {
			if trigger2.ImageChange.LastTriggeredImageID == "" {
				trigger1.ImageChange.LastTriggeredImageID = ""
			}
		}
	}
	if bc2.Spec.SuccessfulBuildsHistoryLimit == nil {
		bc1.Spec.SuccessfulBuildsHistoryLimit = nil
	}
	if bc2.Spec.FailedBuildsHistoryLimit == nil {
		bc1.Spec.FailedBuildsHistoryLimit = nil
	}
	ignoreEmptyMaps(bc1, bc2)

	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{bc1.Name, bc2.Name})
	pairs = append(pairs, [2]interface{}{bc1.Namespace, bc2.Namespace})
	pairs = append(pairs, [2]interface{}{bc1.Labels, bc2.Labels})
	pairs = append(pairs, [2]interface{}{bc1.Annotations, bc2.Annotations})
	pairs = append(pairs, [2]interface{}{bc1.Spec, bc2.Spec})
	equal := EqualPairs(pairs)
	if !equal {
		logger.Info("Resources are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}

func deepEquals(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	struct1 := reflect.ValueOf(deployed).Elem().Type()
	if field1, found1 := struct1.FieldByName("Spec"); found1 {
		struct2 := reflect.ValueOf(requested).Elem().Type()
		if field2, found2 := struct2.FieldByName("Spec"); found2 {
			return Equals(field1, field2)
		}
	}
	return Equals(deployed, requested)
}

func EqualPairs(objects [][2]interface{}) bool {
	for index := range objects {
		if !Equals(objects[index][0], objects[index][1]) {
			return false
		}
	}
	return true
}

func Equals(deployed interface{}, requested interface{}) bool {
	equal := reflect.DeepEqual(deployed, requested)
	if !equal {
		logger.Info("Objects are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}

func ignoreEmptyMaps(deployed metav1.Object, requested metav1.Object) {
	if requested.GetAnnotations() == nil && deployed.GetAnnotations() != nil && len(deployed.GetAnnotations()) == 0 {
		deployed.SetAnnotations(nil)
	}
	if requested.GetLabels() == nil && deployed.GetLabels() != nil && len(deployed.GetLabels()) == 0 {
		deployed.SetLabels(nil)
	}
}
