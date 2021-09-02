module github.com/kiegroup/kogito-operator/client

go 1.16

require (
	github.com/kiegroup/kogito-operator/api v0.0.0
	github.com/stretchr/testify v1.7.0
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	k8s.io/code-generator v0.22.1 // indirect
)

replace github.com/kiegroup/kogito-operator/api => ../api
