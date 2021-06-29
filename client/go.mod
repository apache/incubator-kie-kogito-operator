module github.com/kiegroup/kogito-operator/client

go 1.14

require (
	github.com/kiegroup/kogito-operator/api v0.0.0
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	k8s.io/code-generator v0.21.1 // indirect
)

replace github.com/kiegroup/kogito-operator/api => ../api
