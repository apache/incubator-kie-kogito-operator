module github.com/kiegroup/kogito-cloud-operator

require (
	github.com/RHsyseng/operator-utils v0.0.0-20200304191317-2425bf382482
	github.com/coreos/prometheus-operator v0.40.0
	github.com/cucumber/gherkin-go/v11 v11.0.0
	github.com/cucumber/godog v0.10.0
	github.com/cucumber/messages-go/v10 v10.0.3
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.7
	github.com/gobuffalo/packr/v2 v2.8.0
	github.com/google/uuid v1.1.1
	github.com/imdario/mergo v0.3.8
	github.com/infinispan/infinispan-operator v0.0.0-20200803092941-2b0528367f08
	github.com/integr8ly/grafana-operator/v3 v3.4.0
	github.com/karrick/godirwalk v1.15.6 // indirect
	github.com/keycloak/keycloak-operator v0.0.0-20200917060808-9858b19ca8bf
	github.com/machinebox/graphql v0.2.2
	github.com/matryer/is v1.2.0 // indirect
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/openshift/client-go v0.0.0-20200116152001-92a2713fa240
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20200321030439-57b580e57e88
	github.com/operator-framework/operator-marketplace v0.0.0-20190919183128-4ef67b2f50e9
	github.com/operator-framework/operator-sdk v0.18.2
	github.com/rogpeppe/go-internal v1.6.1 // indirect
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.14.1
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de // indirect
	golang.org/x/sys v0.0.0-20200802091954-4b90ce9b60b3 // indirect
	golang.org/x/tools v0.0.0-20200916195026-c9a70fc28ce3 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.4
	k8s.io/apiextensions-apiserver v0.18.3
	k8s.io/apimachinery v0.18.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	k8s.io/api => k8s.io/api v0.18.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.3
	k8s.io/apiserver => k8s.io/apiserver v0.18.3
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.3
	k8s.io/client-go => k8s.io/client-go v0.18.3
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.18.3
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.3
	k8s.io/code-generator => k8s.io/code-generator v0.18.3
	k8s.io/component-base => k8s.io/component-base v0.18.3
	k8s.io/cri-api => k8s.io/cri-api v0.18.3
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.3
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.3
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.18.3
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.18.3
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.18.3
	k8s.io/kubectl => k8s.io/kubectl v0.18.3
	k8s.io/kubelet => k8s.io/kubelet v0.18.3
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.18.3
	k8s.io/metrics => k8s.io/metrics v0.18.3
	k8s.io/node-api => k8s.io/node-api v0.18.3
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.18.3
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.18.3
	k8s.io/sample-controller => k8s.io/sample-controller v0.18.3
)

replace (
	github.com/openshift/api => github.com/openshift/api v0.0.0-20200623075207-eb651a5bb0ad
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20200623090625-83993cebb5ae
)

// fix Azure bogus https://github.com/kubernetes/client-go/issues/628
replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM

// Required by Helm
replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

go 1.14
