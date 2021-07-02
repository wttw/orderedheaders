package orderedheaders

import (
	"strings"
	"testing"
)

func TestReadMessage(t *testing.T) {
	tests := map[string]struct{
		in string
	}{
		"headersnolf": {"Foo: bar"},
		"headerslf": {"Foo: bar\n"},
		"headeronly": {"Foo: bar\n\n"},
		"emptybody": {"Foo: bar\n\n\n"},
		"withbody": {"Foo: bar\n\nbaz\n"},
	}

	for name, v := range tests {
		t.Run(name, func(t *testing.T){
			_, err := ReadMessage(strings.NewReader(v.in))
			if err != nil {
				t.Fatal("failed to read message", err)
			}
		})
	}
}
