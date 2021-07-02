package orderedheaders

import (
	"io"
	"strings"
	"testing"
)

func TestReadMessage(t *testing.T) {
	tests := map[string]struct{
		in string
		body string
	}{
		"headersnolf": {"Foo: bar", ""},
		"headerslf": {"Foo: bar\n", ""},
		"headeronly": {"Foo: bar\n\n", ""},
		"emptybody": {"Foo: bar\n\n\n", "\n"},
		"withbody": {"Foo: bar\n\nbaz\n", "baz\n"},
	}

	for name, v := range tests {
		t.Run(name, func(t *testing.T){
			msg, err := ReadMessage(strings.NewReader(v.in))
			if err != nil {
				t.Fatal("failed to read message", err)
			}
			if len(msg.Header.Headers) != 1 {
				t.Fatalf("expected one header, got %#v", msg.Header)
			}
			if msg.Header.Headers[0].Key != "Foo" || msg.Header.Headers[0].Value != "bar" {
				t.Fatalf("expected Foo, bar, got %#v", msg.Header.Headers)
			}
			body, err := io.ReadAll(msg.Body)
			if err != nil {
				t.Fatal("failed to read body", err)
			}
			if string(body) != v.body {
				t.Fatalf("body want '%s', got '%s'", v.body, string(body))
			}
		})
	}
}
