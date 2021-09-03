module github.com/kiegroup/kogito-operator/test

go 1.16

require (
	github.com/cucumber/gherkin-go/v11 v11.0.0
	github.com/cucumber/godog v0.11.0
	github.com/cucumber/messages-go/v10 v10.0.3
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.2.0
	github.com/kiegroup/kogito-operator v0.0.0-00010101000000-000000000000
	github.com/kiegroup/kogito-operator/api v0.0.0-00010101000000-000000000000
	github.com/machinebox/graphql v0.2.2
	github.com/matryer/is v1.4.0 // indirect
	github.com/openshift/api v0.0.0-20210105115604-44119421ec6b
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20200321030439-57b580e57e88
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.50.0
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.19.0
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.7
	k8s.io/apiextensions-apiserver v0.20.7
	k8s.io/apimachinery v0.20.7
	k8s.io/client-go v0.20.7
	knative.dev/eventing v0.25.0
	knative.dev/pkg v0.0.0-20210825070025-a70bb26767b8
	sigs.k8s.io/controller-runtime v0.8.3
)

// local modules
replace (
	github.com/RHsyseng/operator-utils => github.com/vaibhavjainwiz/operator-utils v1.4.6-0.20210903063828-d5234ed62c72
	github.com/kiegroup/kogito-operator => ../
	github.com/kiegroup/kogito-operator/api => ../api
)
