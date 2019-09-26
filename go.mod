module github.com/kiegroup/kogito-cloud-operator

require (
	cloud.google.com/go v0.46.3 // indirect
	github.com/NYTimes/gziphandler v1.0.1 // indirect
	github.com/RHsyseng/operator-utils v0.0.0-20190807020041-5344a0f594b8
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.2
	github.com/go-openapi/swag v0.19.5 // indirect
	github.com/gobuffalo/envy v1.7.1 // indirect
	github.com/gobuffalo/logger v1.0.1 // indirect
	github.com/gobuffalo/packr/v2 v2.6.0
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/google/go-cmp v0.3.1
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.11.2 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mailru/easyjson v0.0.0-20190626092158-b2ccc519800e // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo v1.9.0 // indirect
	github.com/onsi/gomega v1.6.0 // indirect
	github.com/openshift/api v0.0.0-00010101000000-000000000000
	github.com/openshift/client-go v0.0.0-00010101000000-000000000000
	github.com/operator-framework/operator-sdk v0.9.1-0.20190712203509-e1d904fa80a4
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/prometheus/client_golang v1.1.0 // indirect
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/prometheus/common v0.7.0 // indirect
	github.com/prometheus/procfs v0.0.5 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	go.opencensus.io v0.22.1 // indirect
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20190911031432-227b76d455e7 // indirect
	golang.org/x/net v0.0.0-20190918130420-a8b05e9114ab // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/sys v0.0.0-20190920190810-ef0ce1748380 // indirect
	golang.org/x/tools v0.0.0-20190920225731-5eefd052ad72 // indirect
	google.golang.org/api v0.10.0 // indirect
	google.golang.org/appengine v1.6.3 // indirect
	google.golang.org/genproto v0.0.0-20190916214212-f660b8655731 // indirect
	google.golang.org/grpc v1.23.1 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190612125737-db0771252981
	k8s.io/apiextensions-apiserver v0.0.0-20190228180357-d002e88f6236
	k8s.io/apimachinery v0.0.0-20190820100751-ac02f8882ef6
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/gengo v0.0.0-20190813173942-955ffa8fcfc9 // indirect
	k8s.io/klog v0.4.0 // indirect
	k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
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

replace github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v0.10.0

replace (
	github.com/openshift/api => github.com/openshift/api v3.9.1-0.20190814194116-a94e914914f4+incompatible
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20190813201236-5a5508328169
)

// thrift moved to github
replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
