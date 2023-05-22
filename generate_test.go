package orderedheaders

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestSet_Singles(t *testing.T) {
	tests := map[string]struct {
		Headers   []KV
		WantError bool
		Want      string
	}{
		"single": {
			[]KV{
				{"subject", "foo"},
			}, false, "Subject: foo\r\n",
		},
		"simple": {
			[]KV{
				{"from", `Steve <steve@blighty.com>`},
				{"to", "bob@example.com"},
				{"Subject", "bar"},
			}, false, "From: \"Steve\" <steve@blighty.com>\r\nTo: <bob@example.com>\r\nSubject: bar\r\n",
		},
		"wrap": {
			[]KV{
				{"subject", "abcdefghi 123456798 abcdefghi 123456798 abcdefghi 123456798 abcdefghi 123456798 abcdefghi 123456798 "},
			}, false, "Subject: abcdefghi 123456798 abcdefghi 123456798 abcdefghi 123456798 abcdefghi\r\n 123456798 abcdefghi 123456798\r\n",
		},
		"long": {
			[]KV{
				{"subject", "abcdefghi123456798abcdefghi123456798abcdefghi123456798abcdefghi123456798abcdefghi 123456798 "},
			}, false, "Subject: abcdefghi123456798abcdefghi123456798abcdefghi123456798abcdefghi123456798abcdefghi\r\n 123456798\r\n",
		},
		"i18n": {
			[]KV{
				{"subject", "Síneadh Fada"},
			}, false, "Subject: =?utf-8?q?S=C3=ADneadh_Fada?=\r\n",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			h := &Header{}
			for _, kv := range test.Headers {
				err := h.Set(kv.Key, kv.Value)
				if err != nil {
					t.Error(err)
					return
				}
			}
			got, err := h.Bytes(Options{})
			if test.WantError {
				t.Errorf("expected error, didn't get one")
			} else {
				if err != nil {
					t.Error(err)
					return
				}
			}
			//fmt.Printf("Got=\n---\n%s\n---\n", string(got))
			if diff := cmp.Diff(test.Want, string(got)); diff != "" {
				t.Errorf("Update mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
