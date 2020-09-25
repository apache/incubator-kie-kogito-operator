package services

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/kafka"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func Test_applyKafkaTopicConfigurations(t *testing.T) {

	appProps := map[string]string{}
	appProps[kafka.QuarkusKafkaBootstrapAppProp] = "kogito-kafka:9092"

	kogitoInfraInstance := &v1alpha1.KogitoInfra{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kogito-kafka",
			Namespace: "mynamespace",
		},
		Spec: v1alpha1.KogitoInfraSpec{
			Resource: v1alpha1.Resource{
				APIVersion: kafka.APIVersion,
				Kind:       kafka.Kind,
			},
		},
	}

	client := test.CreateFakeClient(nil, nil, nil)

	serviceDeployer := serviceDeployer{
		client: client,
		definition: ServiceDefinition{
			KafkaTopics: []KafkaTopicDefinition{
				{TopicName: "kogito-processinstances-events", MessagingType: KafkaTopicIncoming},
			},
		},
	}

	err := serviceDeployer.applyKafkaTopicConfigurations(kogitoInfraInstance, appProps)
	assert.NoError(t, err)
	assert.Equal(t, "kogito-kafka:9092", appProps["mp.messaging.incoming.kogito-processinstances-events.bootstrap.servers"])

	kafkaTopic := &v1beta1.KafkaTopic{}
	exists, err := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Namespace: "mynamespace", Name: "kogito-processinstances-events"}, kafkaTopic)
	assert.NoError(t, err)
	assert.True(t, exists)
}
