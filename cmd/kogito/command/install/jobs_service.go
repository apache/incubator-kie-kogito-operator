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

package install

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/deploy"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	resjobs "github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitojobsservice/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultJobsServiceInfinispanSecretName = "kogito-jobs-service-infinispan-credentials"
	defaultJobsServiceName                 = "jobs-service"
)

type installJobsServiceFlags struct {
	deploy.CommonFlags
	image                         string
	infinispan                    v1alpha1.InfinispanConnectionProperties
	infinispanSasl                string
	infinispanUser                string
	infinispanPassword            string
	backOffRetryMillis            int64
	maxIntervalLimitToRetryMillis int64
	enablePersistence             bool
}

type installJobsServiceCommand struct {
	context.CommandContext
	command *cobra.Command
	flags   installJobsServiceFlags
	Parent  *cobra.Command
}

func newInstallJobsServiceCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := &installJobsServiceCommand{
		CommandContext: *ctx,
		Parent:         parent,
	}

	cmd.RegisterHook()
	cmd.InitHook()

	return cmd
}

func (i *installJobsServiceCommand) Command() *cobra.Command {
	return i.command
}

func (i *installJobsServiceCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "jobs-service [flags]",
		Short:   "Installs the Kogito Jobs Service in the given Project",
		Example: "jobs-service -p my-project",
		Long: `'install jobs-service' deploys the Jobs Service to enable scheduling jobs that aim to be fired at a given time for Kogito Runtime Services.

If 'enable-persistence' flag is set and 'infinispan-url' is not provided, a new Infinispan server will be deployed for you using Kogito Infrastructure.
Use 'infinispan-url' and set 'enable-persistence' flag if you plan to connect to an external Infinispan server that is already provided 
in other namespace or infrastructure.

For more information on Kogito Jobs Service see: https://github.com/kiegroup/kogito-runtimes/wiki/Jobs-Service`,
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
			if len(i.flags.infinispan.URI) > 0 ||
				len(i.flags.infinispanPassword) > 0 ||
				len(i.flags.infinispanUser) > 0 {
				i.flags.enablePersistence = true
			}
			if i.flags.enablePersistence {
				if len(i.flags.infinispan.URI) > 0 {
					i.flags.infinispan.UseKogitoInfra = false
					log.Infof("infinispan-url informed. Infinispan will NOT be provisioned for you. Make sure that %s url is accessible from the cluster", i.flags.infinispan.URI)
				} else {
					if len(i.flags.infinispanPassword) > 0 || len(i.flags.infinispanUser) > 0 {
						return fmt.Errorf("Credentials given, but infinispan-url not set. Please set infinispan URL when providing credentials ")
					}
					i.flags.infinispan.UseKogitoInfra = true
					log.Info("Persistence enabled, Infinispan will be automatically deployed via Infinispan Operator")
				}
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

func (i *installJobsServiceCommand) InitHook() {
	i.flags = installJobsServiceFlags{
		CommonFlags: deploy.CommonFlags{},
		infinispan:  v1alpha1.InfinispanConnectionProperties{},
	}
	i.Parent.AddCommand(i.command)
	deploy.AddDeployFlags(i.command, &i.flags.CommonFlags)

	i.command.Flags().StringVarP(&i.flags.image, "image", "i", resjobs.DefaultImageTagName, "Image tag for the Jobs Service, example: quay.io/kiegroup/kogito-jobs-service:latest")
	i.command.Flags().StringVar(&i.flags.infinispan.URI, "infinispan-url", "", "Set only if enable-persistence is defined. The Infinispan Server URI, example: infinispan-server:11222")
	i.command.Flags().StringVar(&i.flags.infinispan.AuthRealm, "infinispan-authrealm", "", "Set only if enable-persistence is defined. The Infinispan Server Auth Realm for authentication, example: ApplicationRealm")
	i.command.Flags().StringVar(&i.flags.infinispanSasl, "infinispan-sasl", "", "Set only if enable-persistence is defined. The Infinispan Server SASL Mechanism, example: PLAIN")
	i.command.Flags().StringVar(&i.flags.infinispanUser, "infinispan-user", "", "Set only if enable-persistence is defined. The Infinispan Server username")
	i.command.Flags().StringVar(&i.flags.infinispanPassword, "infinispan-password", "", "Set only if enable-persistence is defined. The Infinispan Server password")
	i.command.Flags().BoolVar(&i.flags.enablePersistence, "enable-persistence", false, "Enable persistence using Infinispan. Set also 'infinispan-url' to specify an instance URL. Ifr left in blank the operator will provide one for you")
	i.command.Flags().Int64Var(&i.flags.backOffRetryMillis, "backoff-retry-millis", 0, "Sets the internal property 'kogito.jobs-service.backoffRetryMillis'")
	i.command.Flags().Int64Var(&i.flags.maxIntervalLimitToRetryMillis, "max-internal-limit-retry-millis", 0, "Sets the internal property 'kogito.jobs-service.maxIntervalLimitToRetryMillis'")
}

func (i *installJobsServiceCommand) Exec(cmd *cobra.Command, args []string) error {
	log := context.GetDefaultLogger()
	var err error
	if i.flags.Project, err = shared.EnsureProject(i.Client, i.flags.Project); err != nil {
		return err
	}

	if installed, err := shared.SilentlyInstallOperatorIfNotExists(i.flags.Project, "", i.Client); err != nil {
		return err
	} else if !installed {
		return nil
	}

	if i.flags.infinispan.UseKogitoInfra {
		if available, err := infrastructure.IsInfinispanOperatorAvailable(i.Client, i.flags.Project); err != nil {
			return err
		} else if !available {
			return fmt.Errorf("Infinispan Operator is not available in the Project: %s. Please make sure to install it before deploying Jobs Service without infinispan-url provided ", i.flags.Project)
		}
	}

	// If user and password are sent, create a secret to hold them and attach them to the CRD
	if len(i.flags.infinispanUser) > 0 && len(i.flags.infinispanPassword) > 0 {
		infinispanSecret := v1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: defaultJobsServiceInfinispanSecretName, Namespace: i.flags.Project},
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

	kogitoJobsService := v1alpha1.KogitoJobsService{
		ObjectMeta: metav1.ObjectMeta{Name: defaultJobsServiceName, Namespace: i.flags.Project},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			InfinispanMeta: v1alpha1.InfinispanMeta{
				InfinispanProperties: i.flags.infinispan,
			},
			Replicas: i.flags.Replicas,
			Envs:     shared.FromStringArrayToEnvs(i.flags.Env),
			Image:    shared.FromStringToImage(i.flags.image),
			Resources: v1.ResourceRequirements{
				Limits:   shared.FromStringArrayToResources(i.flags.Limits),
				Requests: shared.FromStringArrayToResources(i.flags.Requests),
			},
			BackOffRetryMillis:            i.flags.backOffRetryMillis,
			MaxIntervalLimitToRetryMillis: i.flags.maxIntervalLimitToRetryMillis,
		},
		Status: v1alpha1.KogitoJobsServiceStatus{
			ConditionsMeta: v1alpha1.ConditionsMeta{Conditions: []v1alpha1.Condition{}},
		},
	}

	if err := kubernetes.ResourceC(i.Client).Create(&kogitoJobsService); err != nil {
		return fmt.Errorf("Error while trying to create a new Kogito Jobs Service: %s ", err)
	}

	log.Infof("Kogito Jobs Service successfully installed in the Project %s.", i.flags.Project)
	log.Infof("Check the Service status by running 'oc describe kogitojobsservice/%s -n %s'", kogitoJobsService.Name, i.flags.Project)

	return nil
}
