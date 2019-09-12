package main

import (
	"reflect"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
)

func Test_fromStringToImage(t *testing.T) {
	type args struct {
		imagetag string
	}
	tests := []struct {
		name string
		args args
		want v1alpha1.Image
	}{
		{"empty", args{""}, v1alpha1.Image{}},
		{"with registry name", args{"quay.io/openshift/myimage:1.0"}, v1alpha1.Image{ImageStreamName: "myimage", ImageStreamTag: "1.0", ImageStreamNamespace: "openshift"}},
		{"full name", args{"openshift/myimage:1.0"}, v1alpha1.Image{ImageStreamName: "myimage", ImageStreamTag: "1.0", ImageStreamNamespace: "openshift"}},
		{"namespace empty", args{"myimage:1.0"}, v1alpha1.Image{ImageStreamName: "myimage", ImageStreamTag: "1.0", ImageStreamNamespace: ""}},
		{"tag empty", args{"myimage"}, v1alpha1.Image{ImageStreamName: "myimage", ImageStreamTag: "", ImageStreamNamespace: ""}},
		{"tag empty with a trick", args{"myimage:"}, v1alpha1.Image{ImageStreamName: "myimage", ImageStreamTag: "", ImageStreamNamespace: ""}},
		{"just tag", args{":1.0"}, v1alpha1.Image{ImageStreamName: "", ImageStreamTag: "1.0", ImageStreamNamespace: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fromStringToImage(tt.args.imagetag); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fromStringToImage() = %v, want %v", got, tt.want)
			}
		})
	}
}
