package orderedheaders

import (
	"bufio"
	"io"
	"net/textproto"
)

type Message struct {
	Header Header
	Body   io.Reader
}

func ReadMessage(r io.Reader) (*Message, error) {
	tp := textproto.NewReader(bufio.NewReader(r))

	hdr, err := ReadHeader(tp)
	if err != nil {
		return nil, err
	}

	return &Message{
		Header: hdr,
		Body:   tp.R,
	}, nil
}
