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

package install

import (
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/deploy"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
)

const (
	defaultDataIndexInfinispanSecretName = infrastructure.DefaultDataIndexName + "-infinispan-credentials"
	defaultInfinispanUsernameKey         = "username"
	defaultInfinispanPasswordKey         = "password"
)

type installDataIndexFlags struct {
	deploy.CommonFlags
	image              string
	httpPort           int32
	kafka              v1alpha1.KafkaConnectionProperties
	infinispan         v1alpha1.InfinispanConnectionProperties
	infinispanSasl     string
	infinispanUser     string
	infinispanPassword string
}

type installDataIndexCommand struct {
	context.CommandContext
	command *cobra.Command
	flags   installDataIndexFlags
	Parent  *cobra.Command
}

func initInstallDataIndexCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := &installDataIndexCommand{
		CommandContext: *ctx,
		Parent:         parent,
	}

	cmd.RegisterHook()
	cmd.InitHook()

	return cmd
}

func (i *installDataIndexCommand) Command() *cobra.Command {
	return i.command
}

func (i *installDataIndexCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "data-index [flags]",
		Short:   "Installs the Kogito Data Index Service in the given Project",
		Example: "data-index -p my-project",
		Long: `'install data-index' will deploy the Data Index service to enable capturing and indexing data produced by one or more Kogito Runtime Services.

If kafka-url is provided, it will be used to connect to the external Kafka server that is deployed in other namespace or infrastructure.
If kafka-instance is provided instead, the value will be used as the Strimzi Kafka instance name to locate the Kafka server deployed in the Data Index service's namespace.
Otherwise, the operator will try to deploy a Kafka instance via Strimzi operator for you using Kogito Infrastructure in the given namespace.

If infinispan-url is not provided, a new Infinispan server will be deployed for you using Kogito Infrastructure, if no one exists in the given project.
Only use infinispan-url if you plan to connect to an external Infinispan server that is already provided in other namespace or infrastructure.

For more information on Kogito Data Index Service see: https://github.com/kiegroup/kogito-runtimes/wiki/Data-Index-Service`,
		RunE:    i.Exec,
		PreRun:  i.CommonPreRun,
		PostRun: i.CommonPostRun,
		Args: func(cmd *cobra.Command, args []string) error {
			log := context.GetDefaultLogger()
			if len(i.flags.infinispanUser) > 0 && len(i.flags.infinispanPassword) == 0 {
				return fmt.Errorf("infinispan-password wasn't provided, please set both infinispan-user and infinispan-password")
			}
			if len(i.flags.infinispanUser) == 0 && len(i.flags.infinispanPassword) > 0 {
				return fmt.Errorf("infinispan-user wasn't provided, please set both infinispan-user and infinispan-password")
			}
			if len(i.flags.infinispanUser) > 0 &&
				len(i.flags.infinispanPassword) > 0 &&
				len(i.flags.infinispanSasl) == 0 {
				i.flags.infinispanSasl = string(v1alpha1.SASLPlain)
			}
			if len(i.flags.infinispan.URI) == 0 {
				i.flags.infinispan.UseKogitoInfra = true
				log.Info("infinispan-url not informed, Infinispan will be automatically deployed via Infinispan Operator")
				if len(i.flags.infinispanPassword) > 0 || len(i.flags.infinispanUser) > 0 {
					return fmt.Errorf("Credentials given, but infinispan-url not set. Please set infinispan URL when providing credentials ")
				}
			} else {
				log.Infof("infinispan-url informed. Infinispan will NOT be provisioned for you. Make sure that %s url is accessible from the cluster", i.flags.infinispan.URI)
				i.flags.infinispan.UseKogitoInfra = false
			}
			if len(i.flags.kafka.ExternalURI) > 0 {
				i.flags.kafka.UseKogitoInfra = false
				log.Infof("kafka-url informed. Kafka will NOT be provisioned for you. Make sure that %s url is accessible from the cluster", i.flags.kafka.ExternalURI)
			} else if len(i.flags.kafka.Instance) > 0 {
				i.flags.kafka.UseKogitoInfra = false
				log.Infof("kafka-instance informed. Kafka will NOT be provisioned for you. Make sure Kafka instance %s is properly deployed in the project. If the Kafka instance is found, Kafka Topics for Data Index service will be deployed in the project if they don't exist already", i.flags.kafka.Instance)
			} else {
				i.flags.kafka.UseKogitoInfra = true
				log.Info("No Kafka information has been given. A Kafka instance will be automatically deployed via Strimzi Operator in the namespace. Kafka Topics will be created accordingly if they don't exist already")
			}
			if err := deploy.CheckDeployArgs(&i.flags.CommonFlags); err != nil {
				return err
			}
			if err := deploy.CheckImageTag(i.flags.image); err != nil {
				return err
			}
			return nil
		},
	}
}

func (i *installDataIndexCommand) InitHook() {
	i.flags = installDataIndexFlags{
		CommonFlags: deploy.CommonFlags{},
		kafka:       v1alpha1.KafkaConnectionProperties{},
		infinispan:  v1alpha1.InfinispanConnectionProperties{},
	}
	i.Parent.AddCommand(i.command)
	deploy.AddDeployFlags(i.command, &i.flags.CommonFlags)

	i.command.Flags().StringVarP(&i.flags.image, "image", "i", "", "Image tag for the Data Index Service, example: quay.io/kiegroup/kogito-data-index:latest")
	i.command.Flags().Int32Var(&i.flags.httpPort, "http-port", framework.DefaultExposedPort, "Default HTTP port which Data Index image will be listening")
	i.command.Flags().StringVar(&i.flags.kafka.ExternalURI, "kafka-url", "", "The Kafka cluster external URI, example: my-kafka-cluster:9092")
	i.command.Flags().StringVar(&i.flags.kafka.Instance, "kafka-instance", "", "The Kafka cluster external URI, example: my-kafka-cluster")
	i.command.Flags().StringVar(&i.flags.infinispan.URI, "infinispan-url", "", "The Infinispan Server URI, example: infinispan-server:11222")
	i.command.Flags().StringVar(&i.flags.infinispan.AuthRealm, "infinispan-authrealm", "", "The Infinispan Server Auth Realm for authentication, example: ApplicationRealm")
	i.command.Flags().StringVar(&i.flags.infinispanSasl, "infinispan-sasl", "", "The Infinispan Server SASL Mechanism, example: PLAIN")
	i.command.Flags().StringVar(&i.flags.infinispanUser, "infinispan-user", "", "The Infinispan Server username")
	i.command.Flags().StringVar(&i.flags.infinispanPassword, "infinispan-password", "", "The Infinispan Server password")
}

func (i *installDataIndexCommand) Exec(cmd *cobra.Command, args []string) error {
	var err error
	if i.flags.Project, err = shared.EnsureProject(i.Client, i.flags.Project); err != nil {
		return err
	}

	// If user and password are sent, create a secret to hold them and attach them to the CRD
	if len(i.flags.infinispanUser) > 0 && len(i.flags.infinispanPassword) > 0 {
		infinispanSecret := v1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: defaultDataIndexInfinispanSecretName, Namespace: i.flags.Project},
		}

		if exist, err := kubernetes.ResourceC(i.Client).Fetch(&infinispanSecret); err != nil {
			return fmt.Errorf("Error while trying to fetch for the Infinispan Credentials Secret: %s ", err)
		} else if exist {
			if err := kubernetes.ResourceC(i.Client).Delete(&infinispanSecret); err != nil {
				return fmt.Errorf("Error while deleting Infinispan Credentials Secret: %s ", err)
			}
		}

		infinispanSecret.StringData = map[string]string{
			defaultInfinispanUsernameKey: i.flags.infinispanUser,
			defaultInfinispanPasswordKey: i.flags.infinispanPassword,
		}

		i.flags.infinispan.Credentials.SecretName = infinispanSecret.Name
		i.flags.infinispan.Credentials.UsernameKey = defaultInfinispanUsernameKey
		i.flags.infinispan.Credentials.PasswordKey = defaultInfinispanPasswordKey
		i.flags.infinispan.UseAuth = true
		i.flags.infinispan.UseKogitoInfra = false
		i.flags.infinispan.SaslMechanism = v1alpha1.InfinispanSaslMechanismType(i.flags.infinispanSasl)
		if err := kubernetes.ResourceC(i.Client).Create(&infinispanSecret); err != nil {
			return fmt.Errorf("Error while trying to create an Infinispan Secret credentials: %s ", err)
		}
	}

	kogitoDataIndex := v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{Name: infrastructure.DefaultDataIndexName, Namespace: i.flags.Project},
		Spec: v1alpha1.KogitoDataIndexSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Replicas: &i.flags.Replicas,
				Envs:     shared.FromStringArrayToEnvs(i.flags.Env),
				Image:    framework.ConvertImageTagToImage(i.flags.image),
				Resources: v1.ResourceRequirements{
					Limits:   shared.FromStringArrayToResources(i.flags.Limits),
					Requests: shared.FromStringArrayToResources(i.flags.Requests),
				},
			},
			HTTPPort:       i.flags.httpPort,
			InfinispanMeta: v1alpha1.InfinispanMeta{InfinispanProperties: i.flags.infinispan},
			KafkaMeta:      v1alpha1.KafkaMeta{KafkaProperties: i.flags.kafka},
		},
		Status: v1alpha1.KogitoDataIndexStatus{
			KogitoServiceStatus: v1alpha1.KogitoServiceStatus{
				ConditionsMeta: v1alpha1.ConditionsMeta{Conditions: []v1alpha1.Condition{}},
			},
		},
	}

	return shared.
		ServicesInstallationBuilder(i.Client, i.flags.Project).
		SilentlyInstallOperatorIfNotExists().
		WarnIfDependenciesNotReady(i.flags.infinispan.UseKogitoInfra, i.flags.kafka.UseKogitoInfra).
		InstallDataIndex(&kogitoDataIndex).
		GetError()
}
