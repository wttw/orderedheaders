package orderedheaders

import (
	"net/mail"
	"net/textproto"
	"regexp"
	"strings"
	"time"
)

// A KV represents a single mime header
type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// A Header represents a MIME-style header consisting
// of a list of key, value pairs
type Header struct {
	Headers []KV `json:"headers"`
}

// ToMap converts a Header to a textproto.MIMEHeader
func (h *Header) ToMap() textproto.MIMEHeader {
	m := make(textproto.MIMEHeader)
	for _, h := range h.Headers {
		m.Add(h.Key, h.Value)
	}
	return m
}

// Add adds a new key, value pair to the header
func (h *Header) Add(key, value string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	h.Headers = append(h.Headers, KV{Key: key, Value: value})
}

// Get gets the first value associated with the given key.
// It is case-insensitive; CanonicalMIMEHeaderKey is used
// to canonicalize the provided key.
// If there are no values associated with the key, Get returns "".
func (h *Header) Get(key string) string {
	key = textproto.CanonicalMIMEHeaderKey(key)
	for _, h := range h.Headers {
		if key == h.Key {
			return h.Value
		}
	}
	return ""
}

// Has returns true if the header specified exists.
func (h *Header) Has(key string) bool {
	key = textproto.CanonicalMIMEHeaderKey(key)
	for _, h := range h.Headers {
		if key == h.Key {
			return true
		}
	}
	return false
}

// AddressList parses the named header field as a list of addresses.
func (h *Header) AddressList(key string) ([]*mail.Address, error) {
	hdr := h.Get(key)
	if hdr == "" {
		return nil, mail.ErrHeaderNotPresent
	}
	return mail.ParseAddressList(hdr)
}

// Date parses the Date header field.
func (h *Header) Date() (time.Time, error) {
	hdr := h.Get("Date")
	if hdr == "" {
		return time.Time{}, mail.ErrHeaderNotPresent
	}
	return mail.ParseDate(hdr)
}

var whitespaceRe = regexp.MustCompile(`[\s\p{Zs}]+`)

// Normalize replaces all whitespace in a header with a single space.
func (h *Header) Normalize() {
	for i, kv := range h.Headers {
		h.Headers[i].Value = strings.TrimSpace(whitespaceRe.ReplaceAllLiteralString(kv.Value, " "))
	}
}

// RemoveAll removes all headers with this (canonicalized) name
func (h *Header) RemoveAll(key string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	filtered := h.Headers[:0]
	for _, kv := range h.Headers {
		if kv.Key != key {
			filtered = append(filtered, kv)
		}
	}
	h.Headers = filtered
}
