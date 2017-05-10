package orderedheaders

import (
	"net/mail"
	"net/textproto"
	"time"
)

// A KV represents a single mime header
type KV struct {
	Key   string
	Value string
}

// A Header represents a MIME-style header consisting
// of a list of key, value pairs
type Header struct {
	Headers []KV
}

// ToMap converts a Header to a textproto.MIMEHeader
func (header *Header) ToMap() textproto.MIMEHeader {
	m := make(textproto.MIMEHeader)
	for _, h := range header.Headers {
		m.Add(h.Key, h.Value)
	}
	return m
}

// Add adds a new key, value pair to the header
func (header *Header) Add(key, value string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	header.Headers = append(header.Headers, KV{Key: key, Value: value})
}

// Get gets the first value associated with the given key.
// It is case insensitive; CanonicalMIMEHeaderKey is used
// to canonicalize the provided key.
// If there are no values associated with the key, Get returns "".
func (header *Header) Get(key string) string {
	key = textproto.CanonicalMIMEHeaderKey(key)
	for _, h := range header.Headers {
		if key == h.Key {
			return h.Value
		}
	}
	return ""
}

// AddressList parses the named header field as a list of addresses.
func (header *Header) AddressList(key string) ([]*mail.Address, error) {
	hdr := header.Get(key)
	if hdr == "" {
		return nil, mail.ErrHeaderNotPresent
	}
	return mail.ParseAddressList(hdr)
}

// Date parses the Date header field.
func (header *Header) Date() (time.Time, error) {
	hdr := header.Get("Date")
	if hdr == "" {
		return time.Time{}, mail.ErrHeaderNotPresent
	}
	return mail.ParseDate(hdr)
}
