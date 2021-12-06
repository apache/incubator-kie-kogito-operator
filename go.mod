module github.com/kiegroup/kogito-operator

go 1.16

require (
	github.com/RHsyseng/operator-utils v1.4.6-0.20210908015233-197f6b3e7a3d
	github.com/go-logr/logr v0.4.0
	github.com/gobuffalo/packr/v2 v2.8.3 // indirect
	github.com/google/uuid v1.3.0
	github.com/kiegroup/kogito-operator/apis v0.0.0-00010101000000-000000000000
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/openshift/api v0.0.0-20210105115604-44119421ec6b
	github.com/openshift/client-go v0.0.0-20210112165513-ebc401615f47
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.50.0
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.19.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914 // indirect
	golang.org/x/sys v0.0.0-20211205182925-97ca703d548d // indirect
	golang.org/x/tools v0.1.8 // indirect
	google.golang.org/grpc v1.39.0 // indirect
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
