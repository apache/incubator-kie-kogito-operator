package resource

import "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"

func fromKafkaToStringMap(kafka v1alpha1.KafkaConnectionProperties) map[string]string {
	propsmap := map[string]string{}
	if &kafka == nil {
		return propsmap
	}

	if len(kafka.ServiceURI) > 0 {
		propsmap[kafkaEnvKeyServiceURI] = kafka.ServiceURI
	}

	return propsmap
}
