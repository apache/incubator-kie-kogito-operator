module github.com/kiegroup/kogito-cloud-operator

require (
	github.com/NYTimes/gziphandler v1.0.1 // indirect
	github.com/RHsyseng/operator-utils v0.0.0-20190807020041-5344a0f594b8
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.2 // indirect
	github.com/go-openapi/swag v0.19.5 // indirect
	github.com/gobuffalo/packr/v2 v2.5.2
	github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d // indirect
	github.com/google/go-cmp v0.2.0
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/json-iterator/go v1.1.7 // indirect
	github.com/mailru/easyjson v0.0.0-20190626092158-b2ccc519800e // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo v1.9.0 // indirect
	github.com/onsi/gomega v1.6.0 // indirect
	github.com/openshift/api v0.0.0-00010101000000-000000000000
	github.com/openshift/client-go v0.0.0-00010101000000-000000000000
	github.com/operator-framework/operator-sdk v0.9.1-0.20190712203509-e1d904fa80a4
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.4.0
	go.uber.org/zap v1.9.1
	golang.org/x/crypto v0.0.0-20190820162420-60c769a6c586 // indirect
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7 // indirect
	golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a // indirect
	golang.org/x/tools v0.0.0-20190820205717-547ecf7b1ef1 // indirect
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190612125737-db0771252981
	k8s.io/apiextensions-apiserver v0.0.0-20190228180357-d002e88f6236
	k8s.io/apimachinery v0.0.0-20190820100751-ac02f8882ef6
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/gengo v0.0.0-20190813173942-955ffa8fcfc9 // indirect
	k8s.io/klog v0.4.0 // indirect
	k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf // indirect
	sigs.k8s.io/controller-runtime v0.1.12
	sigs.k8s.io/controller-tools v0.1.10
)

// Pinned to kubernetes-1.13.4
replace (
	k8s.io/api => k8s.io/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190228180357-d002e88f6236
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190228174230-b40b2a5939e4
)

replace (
	github.com/coreos/prometheus-operator => github.com/coreos/prometheus-operator v0.29.0
	k8s.io/kube-state-metrics => k8s.io/kube-state-metrics v1.6.0
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.1.12
	sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.1.11-0.20190411181648-9d55346c2bde
)

replace github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v0.9.0

replace (
	github.com/openshift/api => github.com/openshift/api v3.9.1-0.20190814194116-a94e914914f4+incompatible
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20190813201236-5a5508328169
)

// thrift moved to github
replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
