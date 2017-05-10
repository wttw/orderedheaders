// Package orderedheaders provides a representation of email
// headers and a way to read them from a textproto.Reader.
package orderedheaders

import (
	"bytes"
	"net/textproto"
)

// ReadHeader reads a MIME-style header from r, much like
// textproto.ReadMIMEHeader.
// The returned value is a list of key, value pairs
func ReadHeader(r *textproto.Reader) (Header, error) {
	m := Header{Headers: []KV{}}
	for {
		kv, err := r.ReadContinuedLineBytes()
		if len(kv) == 0 {
			return m, err
		}
		i := bytes.IndexByte(kv, ':')
		if i < 0 {
			return m, textproto.ProtocolError("malformed MIME header line: " + string(kv))
		}

		endKey := i
		for endKey > 0 && kv[endKey-1] == ' ' {
			endKey--
		}
		key := textproto.CanonicalMIMEHeaderKey(string(kv[:endKey]))
		if key == "" {
			continue
		}

		i++ // colon
		for i < len(kv) && (kv[i] == ' ' || kv[i] == '\t') {
			i++
		}

		value := string(kv[i:])
		m.Add(key, value)
		if err != nil {
			return m, err
		}
	}
}
