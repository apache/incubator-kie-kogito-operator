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
	"fmt"
	infinispan "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	secretName   = "kogito-infinispan-credential"
	replicasSize = 1

	// appPropInfinispanServerList application property for setting infinispan server
	appPropInfinispanServerList int = iota
	// appPropInfinispanUseAuth application property for enabling infinispan authentication
	appPropInfinispanUseAuth
	// appPropInfinispanSaslMechanism application property for setting infinispan SASL mechanism
	// InfinispanSaslMechanismType is the possible SASL Mechanism used during infinispan connection.
	// For more information, see https://en.wikipedia.org/wiki/Simple_Authentication_and_Security_Layer#SASL_mechanisms.
	appPropInfinispanSaslMechanism
	// appPropInfinispanAuthRealm application property for setting infinispan auth realm
	appPropInfinispanAuthRealm
	// envVarInfinispanUser environment variable for setting infinispan username
	envVarInfinispanUser
	// envVarInfinispanPassword environment variable for setting infinispan password
	envVarInfinispanPassword
	infinispanEnvKeyCredSecret = "INFINISPAN_CREDENTIAL_SECRET"
	enablePersistenceEnvKey    = "ENABLE_PERSISTENCE"
	// saslPlain is the PLAIN type.
	saslPlain string = "PLAIN"
)

var (
	//Infinispan variables for the KogitoInfra deployed infrastructure.
	//For Quarkus: https://quarkus.io/guides/infinispan-client#quarkus-infinispan-client_configuration
	//For Spring: https://github.com/infinispan/infinispan-spring-boot/blob/master/infinispan-spring-boot-starter-remote/src/test/resources/test-application.properties

	// propertiesInfinispanQuarkus infinispan properties for quarkus runtime
	propertiesInfinispanQuarkus = map[int]string{
		appPropInfinispanServerList:    "quarkus.infinispan-client.server-list",
		appPropInfinispanUseAuth:       "quarkus.infinispan-client.use-auth",
		appPropInfinispanSaslMechanism: "quarkus.infinispan-client.sasl-mechanism",
		appPropInfinispanAuthRealm:     "quarkus.infinispan-client.auth-realm",

		envVarInfinispanUser:     "QUARKUS_INFINISPAN_CLIENT_AUTH_USERNAME",
		envVarInfinispanPassword: "QUARKUS_INFINISPAN_CLIENT_AUTH_PASSWORD",
	}
	// propertiesInfinispanSpring infinispan properties for spring boot runtime
	propertiesInfinispanSpring = map[int]string{
		appPropInfinispanServerList:    "infinispan.remote.server-list",
		appPropInfinispanUseAuth:       "infinispan.remote.use-auth",
		appPropInfinispanSaslMechanism: "infinispan.remote.sasl-mechanism",
		appPropInfinispanAuthRealm:     "infinispan.remote.auth-realm",

		envVarInfinispanUser:     "INFINISPAN_REMOTE_AUTH_USERNAME",
		envVarInfinispanPassword: "INFINISPAN_REMOTE_AUTH_PASSWORD",
	}
)

func getInfinispanSecretEnvVars(cli *client.Client, infinispanInstance *infinispan.Infinispan, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) ([]corev1.EnvVar, error) {
	var envProps []corev1.EnvVar

	customInfinispanSecret, resultErr := loadCustomKogitoInfinispanSecret(cli, instance.Namespace)
	if resultErr != nil {
		return nil, resultErr
	}

	if customInfinispanSecret == nil {
		customInfinispanSecret, resultErr = createCustomKogitoInfinispanSecret(cli, instance.Namespace, infinispanInstance, instance, scheme)
		if resultErr != nil {
			return nil, resultErr
		}
	}

	envProps = append(envProps, framework.CreateEnvVar(enablePersistenceEnvKey, "true"))
	secretName := customInfinispanSecret.Name
	envProps = append(envProps, framework.CreateEnvVar(infinispanEnvKeyCredSecret, secretName))
	envProps = append(envProps, framework.CreateSecretEnvVar(propertiesInfinispanSpring[envVarInfinispanUser], secretName, infrastructure.InfinispanSecretUsernameKey))
	envProps = append(envProps, framework.CreateSecretEnvVar(propertiesInfinispanQuarkus[envVarInfinispanUser], secretName, infrastructure.InfinispanSecretUsernameKey))
	envProps = append(envProps, framework.CreateSecretEnvVar(propertiesInfinispanSpring[envVarInfinispanPassword], secretName, infrastructure.InfinispanSecretPasswordKey))
	envProps = append(envProps, framework.CreateSecretEnvVar(propertiesInfinispanQuarkus[envVarInfinispanPassword], secretName, infrastructure.InfinispanSecretPasswordKey))
	return envProps, nil
}

func getInfinispanAppProps(cli *client.Client, name string, namespace string) (map[string]string, error) {
	appProps := map[string]string{}

	infinispanURI, resultErr := infrastructure.FetchKogitoInfinispanInstanceURI(cli, name, namespace)
	if resultErr != nil {
		return nil, resultErr
	}

	appProps[propertiesInfinispanSpring[appPropInfinispanUseAuth]] = "true"
	appProps[propertiesInfinispanQuarkus[appPropInfinispanUseAuth]] = "true"
	if len(infinispanURI) > 0 {
		appProps[propertiesInfinispanSpring[appPropInfinispanServerList]] = infinispanURI
		appProps[propertiesInfinispanQuarkus[appPropInfinispanServerList]] = infinispanURI
	}
	appProps[propertiesInfinispanSpring[appPropInfinispanSaslMechanism]] = saslPlain
	appProps[propertiesInfinispanQuarkus[appPropInfinispanSaslMechanism]] = saslPlain
	return appProps, nil
}

func updateInfinispanAppPropsInStatus(cli *client.Client, infinispanInstance *infinispan.Infinispan, instance *v1alpha1.KogitoInfra) error {
	log.Debugf("going to Update Infinispan app properties in kogito infra instance status")
	appProps, err := getInfinispanAppProps(cli, infinispanInstance.Name, infinispanInstance.Namespace)
	if err != nil {
		return err
	}
	instance.Status.AppProps = appProps
	log.Debugf("Following app properties are set infra status : %s", appProps)
	return nil
}

func updateInfinispanEnvVarsInStatus(cli *client.Client, infinispanInstance *infinispan.Infinispan, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) error {
	log.Debugf("going to Update Infinispan env properties in kogito infra instance status")
	envVars, err := getInfinispanSecretEnvVars(cli, infinispanInstance, instance, scheme)
	if err != nil {
		return err
	}
	instance.Status.Env = envVars
	log.Debugf("Following env properties are set infra status : %s", envVars)
	return nil
}

func loadDeployedInfinispanInstance(cli *client.Client, instanceName string, namespace string) (*infinispan.Infinispan, error) {
	log.Debug("fetching deployed kogito infinispan instance")
	infinispanInstance := &infinispan.Infinispan{}
	if exits, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: instanceName, Namespace: namespace}, infinispanInstance); err != nil {
		log.Error("Error occurs while fetching kogito infinispan instance")
		return nil, err
	} else if !exits {
		log.Debug("Kogito infinispan instance is not exists")
		return nil, nil
	} else {
		log.Debug("Kogito infinispan instance found")
		return infinispanInstance, nil
	}
}

func createNewInfinispanInstance(cli *client.Client, name string, namespace string, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) (*infinispan.Infinispan, error) {
	log.Debug("Going to create kogito infinispan instance")
	infinispanRes := &infinispan.Infinispan{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: infinispan.InfinispanSpec{
			Replicas: replicasSize,
		},
	}
	if err := controllerutil.SetOwnerReference(instance, infinispanRes, scheme); err != nil {
		return nil, err
	}
	if err := kubernetes.ResourceC(cli).Create(infinispanRes); err != nil {
		log.Error("Error occurs while creating kogito infinispan instance")
		return nil, err
	}
	log.Debug("Kogito infinispan instance created successfully")
	return infinispanRes, nil
}

func loadCustomKogitoInfinispanSecret(cli *client.Client, namespace string) (*corev1.Secret, error) {
	log.Debugf("Fetching %s ", secretName)
	secret := &corev1.Secret{}
	if exits, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: secretName, Namespace: namespace}, secret); err != nil {
		log.Errorf("Error occurs while fetching %s", secretName)
		return nil, err
	} else if !exits {
		log.Errorf("%s not found", secretName)
		return nil, nil
	} else {
		log.Debugf("%s successfully fetched", secretName)
		return secret, nil
	}
}

func createCustomKogitoInfinispanSecret(cli *client.Client, namespace string, infinispanInstance *infinispan.Infinispan, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) (*corev1.Secret, error) {
	log.Debugf("Creating new secret %s", secretName)

	credentials, err := infrastructure.GetInfinispanCredential(cli, infinispanInstance)
	if err != nil {
		return nil, err
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
	}
	if credentials != nil {
		secret.StringData = map[string]string{
			infrastructure.InfinispanSecretUsernameKey: credentials.Username,
			infrastructure.InfinispanSecretPasswordKey: credentials.Password,
		}
	}
	if err := controllerutil.SetOwnerReference(instance, secret, scheme); err != nil {
		return nil, err
	}
	if err := kubernetes.ResourceC(cli).Create(secret); err != nil {
		log.Errorf("Error occurs while creating %s", secret)
		return nil, err
	}
	log.Debug("%s successfully created", secret)
	return secret, nil
}

type infinispanInfraResource struct {
}

// getInfinispanWatchedObjects provide list of object that needs to be watched to maintain Infinispan kogitoInfra resource
func getInfinispanWatchedObjects() []framework.WatchedObjects {
	return []framework.WatchedObjects{
		{
			GroupVersion: infinispan.SchemeGroupVersion,
			AddToScheme:  infinispan.AddToScheme,
			Objects:      []runtime.Object{&infinispan.Infinispan{}},
		},
		{
			Objects: []runtime.Object{&corev1.Secret{}},
		},
	}
}

// Reconcile reconcile Kogito infra object
func (i *infinispanInfraResource) Reconcile(client *client.Client, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) (requeue bool, resultErr error) {

	var infinispanInstance *infinispan.Infinispan

	if !infrastructure.IsInfinispanAvailable(client) {
		return false, errorForResourceAPINotFound(&instance.Spec.Resource)
	}

	// Step 1: check whether user has provided custom infinispan instance reference
	if len(instance.Spec.Resource.Name) > 0 {
		log.Debugf("Custom infinispan instance reference is provided")

		namespace := instance.Spec.Resource.Namespace
		if len(namespace) == 0 {
			namespace = instance.Namespace
			log.Debugf("Namespace is not provided for custom resource, taking instance namespace(%s) as default", namespace)
		}
		infinispanInstance, resultErr = loadDeployedInfinispanInstance(client, instance.Spec.Resource.Name, namespace)
		if resultErr != nil {
			return false, resultErr
		}
		if infinispanInstance == nil {
			return false, errorForResourceNotFound("Infinispan", instance.Spec.Resource.Name, namespace)
		}
	} else {
		// create/refer kogito-infinispan instance
		log.Debugf("Custom infinispan instance reference is not provided")
		// Verify Infinispan Operator (it's installation is required in the same namespace, that's why we do this check as well)
		if infinispanAvailable, err := infrastructure.IsInfinispanOperatorAvailable(client, instance.Namespace); err != nil {
			return false, err
		} else if !infinispanAvailable {
			return false, errorForResourceAPINotFound(&instance.Spec.Resource)
		}
		infinispanInstance, resultErr = loadDeployedInfinispanInstance(client, infrastructure.InfinispanInstanceName, instance.Namespace)
		if resultErr != nil {
			return false, resultErr
		}

		if infinispanInstance == nil {
			// if not exist then create new Infinispan instance. Infinispan operator creates Infinispan instance, secret & service resource
			_, resultErr = createNewInfinispanInstance(client, infrastructure.InfinispanInstanceName, instance.Namespace, instance, scheme)
			if resultErr != nil {
				return false, resultErr
			}
			return true, nil
		}
	}
	infinispanCondition := getLatestInfinispanCondition(infinispanInstance)
	if infinispanCondition == nil || infinispanCondition.Status != string(v1.ConditionTrue) {
		return false, errorForResourceNotReadyError(fmt.Errorf("infinispan instance %s not ready. Waiting for Condition.Type == True", infinispanInstance.Name))
	}
	if resultErr := updateInfinispanAppPropsInStatus(client, infinispanInstance, instance); resultErr != nil {
		return false, nil
	}
	if resultErr := updateInfinispanEnvVarsInStatus(client, infinispanInstance, instance, scheme); resultErr != nil {
		return false, nil
	}
	return false, resultErr
}

func getLatestInfinispanCondition(instance *infinispan.Infinispan) *infinispan.InfinispanCondition {
	if len(instance.Status.Conditions) == 0 {
		return nil
	}
	// Infinispan does not have a condition transition date, so let's get the latest
	return &instance.Status.Conditions[len(instance.Status.Conditions)-1]
}
