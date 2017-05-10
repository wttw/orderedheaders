[![GoDoc](https://godoc.org/github.com/wttw/orderedheaders?status.svg)](https://godoc.org/github.com/wttw/orderedheaders)

# orderedheaders
Header parsing retaining order in Go

[textproto](https://golang.org/pkg/net/textproto/) includes
a very nice MIME header parser. But it returns all the
headers as a map, meaning it loses any ordering information.

For MIME headers that's seldom a problem, but for email headers
it can occasionally be.

This package provides an alternate representation of headers
read from a textproto reader as a list of key, value pairs.
It also includes a few helper functions that are compatible
with those in the [textproto](https://golang.org/pkg/net/textproto/)
and [mail](https://golang.org/pkg/net/mail/) packages.