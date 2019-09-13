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

package main

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	resdataindex "github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitodataindex/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
)

type installDataIndexFlags struct {
	deployCommonFlags
	image              string
	kafka              v1alpha1.KafkaConnectionProperties
	infinispan         v1alpha1.InfinispanConnectionProperties
	infinispanSasl     string
	infinispanUser     string
	infinispanPassword string
}

const (
	defaultDataIndexName         = "kogito-data-index"
	defaultInfinispanSecretName  = "kogito-data-index-infinispan-credentials"
	defaulInfinispanUsernameKey  = "username"
	defaultInfinispanPasswordKey = "password"
)

var (
	installDataIndexCmd      *cobra.Command
	installDataIndexCmdFlags = installDataIndexFlags{
		deployCommonFlags: deployCommonFlags{},
		kafka:             v1alpha1.KafkaConnectionProperties{},
		infinispan:        v1alpha1.InfinispanConnectionProperties{},
	}
)

var _ = RegisterCommandVar(func() {
	installDataIndexCmd = &cobra.Command{
		Use:     "data-index [flags]",
		Short:   "Installs the Kogito Data Index Service in the given Project",
		Example: "data-index -p my-project --infinispan-url my-infinispan-server:11222 --kafka-url my-kafka-bootstrap:9092",
		Long: `'install data-index' will deploy the Data Index service to enable capturing and indexing data produced by one or more Kogito Runtime Services.

				kafka-url and infinispan-url are required options since the Data Index Service needs both servers to be up and running in the cluster to 
                work correctly. Please refer to the https://github.com/kiegroup/kogito-cloud-operator#install-data-index-service for more information regarding
                how to deploy an Infinispan and Kafka cluster on OpenShift for the Data Index to use.

				For more information on Kogito Data Index Service see: https://github.com/kiegroup/kogito-runtimes/wiki/Data-Index-Service`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return installDataIndexExec(cmd, args)
		},
		PreRun:  preRunF,
		PostRun: posRunF,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(installDataIndexCmdFlags.infinispanUser) > 0 && len(installDataIndexCmdFlags.infinispanPassword) == 0 {
				return fmt.Errorf("infinispan-password wasn't provided, please set both infinispan-user and infinispan-password")
			}
			if len(installDataIndexCmdFlags.infinispanUser) == 0 && len(installDataIndexCmdFlags.infinispanPassword) > 0 {
				return fmt.Errorf("infinispan-user wasn't provided, please set both infinispan-user and infinispan-password")
			}
			if len(installDataIndexCmdFlags.infinispanUser) > 0 &&
				len(installDataIndexCmdFlags.infinispanPassword) > 0 &&
				len(installDataIndexCmdFlags.infinispanSasl) == 0 {
				installDataIndexCmdFlags.infinispanSasl = string(v1alpha1.SASLPlain)
			}
			if err := commonCheckDeployArgs(&installDataIndexCmdFlags.deployCommonFlags); err != nil {
				return err
			}
			if err := commonCheckImageTag(installDataIndexCmdFlags.image); err != nil {
				return err
			}
			return nil
		},
	}
})

var _ = RegisterCommandInit(func() {
	installCmd.AddCommand(installDataIndexCmd)
	commonAddDeployFlags(installDataIndexCmd, &installDataIndexCmdFlags.deployCommonFlags)

	installDataIndexCmd.Flags().StringVarP(&installDataIndexCmdFlags.image, "image", "i", resdataindex.DefaultImage, "Image tag (namespace/name:tag) for the runtime Service, e.g: openshift/kogito-data-index:latest")
	installDataIndexCmd.Flags().StringVar(&installDataIndexCmdFlags.kafka.ServiceURI, "kafka-url", "", "The Kafka cluster internal URI, example: my-kafka-cluster:9092")
	installDataIndexCmd.Flags().StringVar(&installDataIndexCmdFlags.infinispan.ServiceURI, "infinispan-url", "", "The Infinispan Server internal URI, example: infinispan-server:11222")
	installDataIndexCmd.Flags().StringVar(&installDataIndexCmdFlags.infinispan.AuthRealm, "infinispan-authrealm", "", "The Infinispan Server Auth Realm for authentication, example: ApplicationRealm")
	installDataIndexCmd.Flags().StringVar(&installDataIndexCmdFlags.infinispanSasl, "infinispan-sasl", "", "The Infinispan Server SASL Mechanism, example: PLAIN")
	installDataIndexCmd.Flags().StringVar(&installDataIndexCmdFlags.infinispanUser, "infinispan-user", "", "The Infinispan Server username")
	installDataIndexCmd.Flags().StringVar(&installDataIndexCmdFlags.infinispanPassword, "infinispan-password", "", "The Infinispan Server password")

	cobra.MarkFlagRequired(installDataIndexCmd.Flags(), "kafka-url")
	cobra.MarkFlagRequired(installDataIndexCmd.Flags(), "infinispan-url")
})

func installDataIndexExec(cmd *cobra.Command, args []string) error {
	var err error
	if installDataIndexCmdFlags.project, err = ensureProject(installDataIndexCmdFlags.project); err != nil {
		return err
	}
	if err := checkKogitoDataIndexCRDExists(); err != nil {
		return err
	}

	// if user and password is sent, let's create a secret to hold it and attach to the CRD
	if len(installDataIndexCmdFlags.infinispanUser) > 0 && len(installDataIndexCmdFlags.infinispanPassword) > 0 {
		infinispanSecret := v1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: defaultInfinispanSecretName, Namespace: installDataIndexCmdFlags.project},
		}

		if exist, err := kubernetes.ResourceC(kubeCli).Fetch(&infinispanSecret); err != nil {
			return fmt.Errorf("Error while trying to fetch for the Infinispan Credentials Secret: %s", err)
		} else if exist {
			if err := kubernetes.ResourceC(kubeCli).Delete(&infinispanSecret); err != nil {
				return fmt.Errorf("Error while deleting Infinispan Credentials Secret: %s", err)
			}
		}

		infinispanSecret.StringData = map[string]string{
			defaulInfinispanUsernameKey:  installDataIndexCmdFlags.infinispanUser,
			defaultInfinispanPasswordKey: installDataIndexCmdFlags.infinispanPassword,
		}

		installDataIndexCmdFlags.infinispan.Credentials.SecretName = infinispanSecret.Name
		installDataIndexCmdFlags.infinispan.Credentials.UsernameKey = defaulInfinispanUsernameKey
		installDataIndexCmdFlags.infinispan.Credentials.PasswordKey = defaultInfinispanPasswordKey
		installDataIndexCmdFlags.infinispan.UseAuth = true
		installDataIndexCmdFlags.infinispan.SaslMechanism = v1alpha1.InfinispanSaslMechanismType(installDataIndexCmdFlags.infinispanSasl)
		if err := kubernetes.ResourceC(kubeCli).Create(&infinispanSecret); err != nil {
			return fmt.Errorf("Error while trying to create an Infinispan Secret credentials: %s", err)
		}
	}

	kogitoDataIndex := v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{Name: defaultDataIndexName, Namespace: installDataIndexCmdFlags.project},
		Spec: v1alpha1.KogitoDataIndexSpec{
			Name:          defaultDataIndexName,
			Replicas:      installDataIndexCmdFlags.replicas,
			Env:           util.FromStringsKeyPairToMap(installDataIndexCmdFlags.env),
			Image:         installDataIndexCmdFlags.image,
			MemoryLimit:   extractResource(v1alpha1.ResourceMemory, installDataIndexCmdFlags.limits),
			MemoryRequest: extractResource(v1alpha1.ResourceMemory, installDataIndexCmdFlags.requests),
			CPULimit:      extractResource(v1alpha1.ResourceCPU, installDataIndexCmdFlags.limits),
			CPURequest:    extractResource(v1alpha1.ResourceCPU, installDataIndexCmdFlags.requests),
			Infinispan:    installDataIndexCmdFlags.infinispan,
			Kafka:         installDataIndexCmdFlags.kafka,
		},
	}

	if err := kubernetes.ResourceC(kubeCli).Create(&kogitoDataIndex); err != nil {
		return fmt.Errorf("Error while trying to create a new Kogito Data Index Service: %s", err)
	}

	log.Infof("Kogito Data Index Service successfully installed in the Project %s.", installDataIndexCmdFlags.project)
	log.Infof("Check the Service status by running 'oc describe kogitodataindex/%s -n %s'", kogitoDataIndex.Spec.Name, installDataIndexCmdFlags.project)

	return nil
}
