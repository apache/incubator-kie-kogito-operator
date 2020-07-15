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

package mappers

import (
	"fmt"
	"strconv"

	"github.com/cucumber/messages-go/v10"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/test/types"
	bddtypes "github.com/kiegroup/kogito-cloud-operator/test/types"
)

const (
	enabledKey  = "enabled"
	disabledKey = "disabled"

	// DataTable first column
	tableRowConfigKey          = "config"
	tableRowWebhookKey         = "webhook"
	tableRowInfinispanKey      = "infinispan"
	tableRowKafkaKey           = "kafka"
	tableRowRuntimeRequestKey  = "runtime-request"
	tableRowRuntimeLimitKey    = "runtime-limit"
	tableRowRuntimeEnvKey      = "runtime-env"
	tableRowServiceLabelKey    = "service-label"
	tableRowDeploymentLabelKey = "deployment-label"

	// DataTable second column
	tableRowNativeKey                   = "native"
	tableRowInfinispanUseKogitoInfraKey = "useKogitoInfra"
	tableRowInfinispanUsernameKey       = "username"
	tableRowInfinispanPasswordKey       = "password"
	tableRowInfinispanURIKey            = "uri"
	tableRowKafkaUseKogitoInfraKey      = "useKogitoInfra"
	tableRowKafkaExternalURIKey         = "externalURI"
	tableRowKafkaInstanceKey            = "instance"
	tableRowHTTPPortKey                 = "httpPort"
	tableRowTypeKey                     = "type"
	tableRowSecretKey                   = "secret"
)

// MapKogitoServiceTable maps Cucumber table row to KogitoBuildHolder
func MapKogitoServiceTable(table *messages.PickleStepArgument_PickleTable, serviceHolder *types.KogitoServiceHolder) error {
	for _, row := range table.Rows {
		// Try to map configuration row to KogitoServiceHolder
		_, err := mapKogitoServiceTableRow(row, serviceHolder)
		if err != nil {
			return err
		}

	}
	return nil
}

// mapKogitoServiceTableRow maps Cucumber table row to KogitoServiceHolder
func mapKogitoServiceTableRow(row *messages.PickleStepArgument_PickleTable_PickleTableRow, kogitoService *bddtypes.KogitoServiceHolder) (mappingFound bool, err error) {
	if len(row.Cells) != 3 {
		return false, fmt.Errorf("expected table to have exactly three columns")
	}

	firstColumn := getFirstColumn(row)

	switch firstColumn {
	case tableRowInfinispanKey:
		return mapKogitoServiceInfinispanTableRow(row, kogitoService)

	case tableRowKafkaKey:
		return mapKogitoServiceKafkaTableRow(row, kogitoService)

	case tableRowServiceLabelKey:
		kogitoService.KogitoService.GetSpec().GetServiceLabels()[getSecondColumn(row)] = getThirdColumn(row)

	case tableRowDeploymentLabelKey:
		kogitoService.KogitoService.GetSpec().GetDeploymentLabels()[getSecondColumn(row)] = getThirdColumn(row)

	case tableRowRuntimeEnvKey:
		kogitoService.KogitoService.GetSpec().AddEnvironmentVariable(getSecondColumn(row), getThirdColumn(row))

	case tableRowRuntimeRequestKey:
		kogitoService.KogitoService.GetSpec().AddResourceRequest(getSecondColumn(row), getThirdColumn(row))

	case tableRowRuntimeLimitKey:
		kogitoService.KogitoService.GetSpec().AddResourceLimit(getSecondColumn(row), getThirdColumn(row))

	case tableRowConfigKey:
		return mapKogitoServiceConfigTableRow(row, kogitoService)

	default:
		return false, fmt.Errorf("Unrecognized configuration option: %s", firstColumn)
	}

	return true, nil
}

func mapKogitoServiceInfinispanTableRow(row *messages.PickleStepArgument_PickleTable_PickleTableRow, kogitoService *bddtypes.KogitoServiceHolder) (mappingFound bool, err error) {
	secondColumn := getSecondColumn(row)

	if infinispanAware, ok := kogitoService.KogitoService.GetSpec().(v1alpha1.InfinispanAware); ok {
		switch secondColumn {
		case tableRowInfinispanUseKogitoInfraKey:
			infinispanAware.GetInfinispanProperties().UseKogitoInfra = mustParseEnabledDisabled(getThirdColumn(row))

		case tableRowInfinispanUsernameKey:
			kogitoService.Infinispan.Username = getThirdColumn(row)

		case tableRowInfinispanPasswordKey:
			kogitoService.Infinispan.Password = getThirdColumn(row)

		case tableRowInfinispanURIKey:
			infinispanAware.GetInfinispanProperties().URI = getThirdColumn(row)

		default:
			return false, fmt.Errorf("Unrecognized Infinispan configuration option: %s", secondColumn)
		}
	} else {
		return false, fmt.Errorf("Kogito service %s doesn't support Infinispan configuration", kogitoService.KogitoService.GetName())
	}
	return true, nil
}

func mapKogitoServiceKafkaTableRow(row *messages.PickleStepArgument_PickleTable_PickleTableRow, kogitoService *bddtypes.KogitoServiceHolder) (mappingFound bool, err error) {
	secondColumn := getSecondColumn(row)

	if kafkaAware, ok := kogitoService.KogitoService.GetSpec().(v1alpha1.KafkaAware); ok {
		switch secondColumn {
		case tableRowKafkaUseKogitoInfraKey:
			kafkaAware.GetKafkaProperties().UseKogitoInfra = mustParseEnabledDisabled(getThirdColumn(row))

		case tableRowKafkaExternalURIKey:
			kafkaAware.GetKafkaProperties().ExternalURI = getThirdColumn(row)

		case tableRowKafkaInstanceKey:
			kafkaAware.GetKafkaProperties().Instance = getThirdColumn(row)

		default:
			return false, fmt.Errorf("Unrecognized Kafka configuration option: %s", secondColumn)
		}
	} else {
		return false, fmt.Errorf("Kogito service %s doesn't support Kafka configuration", kogitoService.KogitoService.GetName())
	}
	return true, nil
}

func mapKogitoServiceConfigTableRow(row *messages.PickleStepArgument_PickleTable_PickleTableRow, kogitoService *bddtypes.KogitoServiceHolder) (mappingFound bool, err error) {
	secondColumn := getSecondColumn(row)

	switch secondColumn {
	case tableRowHTTPPortKey:
		httpPort, err := strconv.ParseInt(getThirdColumn(row), 10, 32)
		if err != nil {
			return false, err
		}

		kogitoService.KogitoService.GetSpec().SetHTTPPort(int32(httpPort))

	default:
		return false, fmt.Errorf("Unrecognized config configuration option: %s", secondColumn)
	}

	return true, nil
}

// MapKogitoBuildTable maps Cucumber table to KogitoBuildHolder
func MapKogitoBuildTable(table *messages.PickleStepArgument_PickleTable, buildHolder *types.KogitoBuildHolder) error {
	for _, row := range table.Rows {
		// Try to map configuration row to KogitoServiceHolder
		mappingFound, serviceMapErr := mapKogitoServiceTableRow(row, buildHolder.KogitoServiceHolder)
		if !mappingFound {
			// Try to map configuration row to KogitoBuild
			mappingFound, buildMapErr := mapKogitoBuildTableRow(row, buildHolder.KogitoBuild)
			if !mappingFound {
				return fmt.Errorf("Row mapping not found, Kogito service mapping error: %v , Kogito build mapping error: %v", serviceMapErr, buildMapErr)
			}
		}

	}
	return nil
}

// mapKogitoBuildTableRow maps Cucumber table row to KogitoBuild
func mapKogitoBuildTableRow(row *messages.PickleStepArgument_PickleTable_PickleTableRow, kogitoBuild *v1alpha1.KogitoBuild) (mappingFound bool, err error) {
	if len(row.Cells) != 3 {
		return false, fmt.Errorf("expected table to have exactly three columns")
	}

	firstColumn := getFirstColumn(row)

	switch firstColumn {
	case tableRowConfigKey:
		return mapKogitoBuildConfigTableRow(row, kogitoBuild)

	case tableRowWebhookKey:
		return mapKogitoBuildWebhookTableRow(row, kogitoBuild)

	default:
		return false, fmt.Errorf("Unrecognized configuration option: %s", firstColumn)
	}
}

func mapKogitoBuildConfigTableRow(row *messages.PickleStepArgument_PickleTable_PickleTableRow, kogitoBuild *v1alpha1.KogitoBuild) (mappingFound bool, err error) {
	secondColumn := getSecondColumn(row)

	switch secondColumn {
	case tableRowNativeKey:
		kogitoBuild.Spec.Native = mustParseEnabledDisabled(getThirdColumn(row))

	default:
		return false, fmt.Errorf("Unrecognized config configuration option: %s", secondColumn)
	}

	return true, nil
}

func mapKogitoBuildWebhookTableRow(row *messages.PickleStepArgument_PickleTable_PickleTableRow, kogitoBuild *v1alpha1.KogitoBuild) (mappingFound bool, err error) {
	secondColumn := getSecondColumn(row)

	if len(kogitoBuild.Spec.WebHooks) == 0 {
		kogitoBuild.Spec.WebHooks = []v1alpha1.WebhookSecret{{}}
	}

	switch secondColumn {
	case tableRowTypeKey:
		kogitoBuild.Spec.WebHooks[0].Type = v1alpha1.WebhookType(getThirdColumn(row))
	case tableRowSecretKey:
		kogitoBuild.Spec.WebHooks[0].Secret = getThirdColumn(row)

	default:
		return false, fmt.Errorf("Unrecognized webhook configuration option: %s", secondColumn)
	}

	return true, nil
}

// Helper methods

func getFirstColumn(row *messages.PickleStepArgument_PickleTable_PickleTableRow) string {
	return row.Cells[0].Value
}

func getSecondColumn(row *messages.PickleStepArgument_PickleTable_PickleTableRow) string {
	return row.Cells[1].Value
}

func getThirdColumn(row *messages.PickleStepArgument_PickleTable_PickleTableRow) string {
	return row.Cells[2].Value
}

// mustParseEnabledDisabled parse a boolean string value
func mustParseEnabledDisabled(value string) bool {
	switch value {
	case enabledKey:
		return true
	case disabledKey:
		return false
	default:
		panic(fmt.Errorf("Unknown value for enabled/disabled: %s", value))
	}
}
