module github.com/kiegroup/kogito-operator

go 1.16

require (
	cloud.google.com/go v0.99.0 // indirect
	github.com/RHsyseng/operator-utils v1.4.6-0.20210908015233-197f6b3e7a3d
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-logr/logr v0.4.0
	github.com/google/uuid v1.3.0
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kiegroup/kogito-operator/apis v0.0.0-00010101000000-000000000000
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/openshift/api v0.0.0-20210105115604-44119421ec6b
	github.com/openshift/client-go v0.0.0-20210112165513-ebc401615f47
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.50.0
	github.com/spf13/cobra v1.5.0
	github.com/stretchr/testify v1.8.0
	go.uber.org/zap v1.19.0
	golang.org/x/sync v0.0.0-20220819030929-7fc1605a5dde // indirect
	golang.org/x/sys v0.0.0-20220829200755-d48e67d00261 // indirect
	golang.org/x/term v0.0.0-20220722155259-a9ba230a4035 // indirect
	golang.org/x/tools v0.1.12 // indirect
	google.golang.org/api v0.62.0 // indirect
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa // indirect
	google.golang.org/grpc v1.42.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.7
	k8s.io/apiextensions-apiserver v0.20.7
	k8s.io/apimachinery v0.20.7
	k8s.io/client-go v0.20.7
	knative.dev/eventing v0.25.0
	knative.dev/pkg v0.0.0-20210825070025-a70bb26767b8
	sigs.k8s.io/controller-runtime v0.8.3
	software.sslmate.com/src/go-pkcs12 v0.0.0-20210415151418-c5206de65a78
)

// local modules
replace github.com/kiegroup/kogito-operator/apis => ./apis
