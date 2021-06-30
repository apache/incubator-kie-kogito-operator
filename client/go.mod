module github.com/kiegroup/kogito-operator/client

go 1.14

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/kiegroup/kogito-operator/api v0.0.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20210224082022-3d97a244fca7 // indirect
	golang.org/x/sys v0.0.0-20210225134936-a50acf3fe073 // indirect
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	k8s.io/code-generator v0.21.1 // indirect
	k8s.io/klog/v2 v2.8.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.0 // indirect
)

replace github.com/kiegroup/kogito-operator/api => ../api
