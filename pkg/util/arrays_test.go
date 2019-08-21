package util

import (
	"reflect"
	"testing"
)

func TestFromStringsKeyPairToMap(t *testing.T) {
	type args struct {
		array []string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{"happy path", args{[]string{"env1=value", "env2=value"}}, map[string]string{"env1": "value", "env2": "value"}},
		{"happy path2", args{[]string{"env1=value"}}, map[string]string{"env1": "value"}},
		{"only key", args{[]string{"env1="}}, map[string]string{"env1": ""}},
		{"only key without sep", args{[]string{"env1"}}, map[string]string{"env1": ""}},
		{"no key no value", args{[]string{""}}, map[string]string{}},
		{"no key with value", args{[]string{"=value"}}, map[string]string{}},
		{"various seps", args{[]string{"env1=value=value"}}, map[string]string{"env1": "value=value"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromStringsKeyPairToMap(tt.args.array); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromStringsKeyPairToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
