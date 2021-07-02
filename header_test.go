package orderedheaders

import "testing"

func TestHeaderNormalize(t *testing.T) {
	in := Header{
		Headers: []KV{
			{"FOO", "  one\ttwo   three\n  four\t five\t\nsix\r\nseven\n"},
		},
	}
	in.Normalize()
	want := "one two three four five six seven"
	got := in.Headers[0].Value
	if got != want {
		t.Errorf("want: '%s', got: '%s'", want, got)
	}
}
