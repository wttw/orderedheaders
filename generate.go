package orderedheaders

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/mail"
	"net/textproto"
	"regexp"
	"strings"
)

//go:generate enumer -json -trimprefix=HeaderType -transform=kebab -type HeaderType

// Functions to help generate valid email headers, outside the base
// textproto.MIMEHeaders equivalent functionality.

// It is opinionated, and will attempt to fix up invalid user input
// when possible

const (
	HdrReturnPath              = "Return-Path"
	HdrReceived                = "Received"
	HdrDate                    = "Date"
	HdrFrom                    = "From"
	HdrSender                  = "Sender"
	HdrReplyTo                 = "Reply-To"
	HdrTo                      = "To"
	HdrCc                      = "Cc"
	HdrBcc                     = "Bcc"
	HdrMessageId               = "Message-Id"
	HdrInReplyTo               = "In-Reply-To"
	HdrReferences              = "References"
	HdrSubject                 = "Subject"
	HdrComments                = "Comments"
	HdrKeywords                = "Keywords"
	HdrResentDate              = "Resent-Date"
	HdrResentFrom              = "Resent-From"
	HdrResentSender            = "Resent-Sender"
	HdrResentTo                = "Resent-To"
	HdrResentCc                = "Resent-Cc"
	HdrResentBcc               = "Resent-Bcc"
	HdrMimeVersion             = "Mime-Version"
	HdrContentType             = "Content-Type"
	HdrContentID               = "Content-ID"
	HdrContentTransferEncoding = "Content-Transfer-Encoding"
	HdrContentDescription      = "Content-Description"
)

const utf8 = "utf-8"

// HeaderType describes the required syntax for an email header
type HeaderType int

const (
	HeaderTypeUnstructured HeaderType = iota
	HeaderTypeMailbox
	HeaderTypeMailboxList
	HeaderTypeDate
	HeaderTypeReceived
	HeaderTypeMessageID
	HeaderTypeMessageIDList
	HeaderTypePhraseList
	HeaderTypeReturnPath
	HeaderTypeOpaque
)

// Syntax contains RFC 5322 requirements for a header
type Syntax struct {
	Required bool
	Unique   bool
	Type     HeaderType
}

// https://tools.wordtothewise.com/rfc5322#section-3.6
// https://tools.wordtothewise.com/rfc2045#section-3

// HeaderSyntax maps header names to their syntax
var HeaderSyntax = map[string]Syntax{
	HdrReturnPath:              {Type: HeaderTypeReturnPath},
	HdrReceived:                {Type: HeaderTypeReceived},
	HdrDate:                    {Required: true, Unique: true, Type: HeaderTypeDate},
	HdrFrom:                    {Required: true, Unique: true, Type: HeaderTypeMailboxList},
	HdrSender:                  {Unique: true, Type: HeaderTypeMailbox},
	HdrReplyTo:                 {Unique: true, Type: HeaderTypeMailboxList},
	HdrTo:                      {Unique: true, Type: HeaderTypeMailboxList},
	HdrCc:                      {Unique: true, Type: HeaderTypeMailboxList},
	HdrBcc:                     {Unique: true, Type: HeaderTypeMailboxList},
	HdrMessageId:               {Unique: true, Type: HeaderTypeMessageID},
	HdrInReplyTo:               {Unique: true, Type: HeaderTypeMessageIDList},
	HdrReferences:              {Unique: true, Type: HeaderTypeMessageIDList},
	HdrSubject:                 {Unique: true, Type: HeaderTypeUnstructured},
	HdrComments:                {Type: HeaderTypeUnstructured},
	HdrKeywords:                {Type: HeaderTypePhraseList},
	HdrResentDate:              {Type: HeaderTypeDate},
	HdrResentFrom:              {Type: HeaderTypeMailboxList},
	HdrResentSender:            {Type: HeaderTypeMailbox},
	HdrResentTo:                {Type: HeaderTypeMailboxList},
	HdrResentCc:                {Type: HeaderTypeMailboxList},
	HdrResentBcc:               {Type: HeaderTypeMailboxList},
	HdrMimeVersion:             {Unique: true, Type: HeaderTypeOpaque},
	HdrContentType:             {Unique: true, Type: HeaderTypeOpaque},
	HdrContentID:               {Unique: true, Type: HeaderTypeMessageID},
	HdrContentTransferEncoding: {Unique: true, Type: HeaderTypeOpaque},
	HdrContentDescription:      {Unique: true, Type: HeaderTypeUnstructured},
}

// Options configures how a set of headers will be rendered.
type Options struct {
	// RenderBCC enables rendering the Bcc: header, which is ignored by default
	RenderBCC bool
	// RenderBlank enables rendering headers which have zero length content
	RenderBlank bool
	// NoEscape disables encoding of non-ASCI content in a header
	NoEscape bool
	// RenderReturnPath enables rendering the Return-Path: header, which is ignored by default
	RenderReturnPath bool
}

// Set sets a standard header, replacing any existing one. It only accepts
// standard email headers, not extensions.
func (h *Header) Set(key, value string) error {
	canonKey := textproto.CanonicalMIMEHeaderKey(key)
	syntax, ok := HeaderSyntax[canonKey]
	if !ok {
		return fmt.Errorf("%s is not a standard email header", canonKey)
	}
	if value != "" {
		err := checkHeader(syntax.Type, value)
		if err != nil {
			return fmt.Errorf("invalid value for %s: %w", key, err)
		}
	}
	for i, v := range h.Headers {
		if v.Key == canonKey {
			h.Headers[i] = KV{
				Key:   canonKey,
				Value: value,
			}
			return nil
		}
	}
	h.Headers = append(h.Headers, KV{
		Key:   canonKey,
		Value: value,
	})
	return nil
}

func (h *Header) WriteTo(w io.Writer, o Options) error {
	seen := map[string]struct{}{}
	for _, h := range h.Headers {
		if !o.RenderBlank && h.Value == "" {
			continue
		}
		if h.Key == HdrBcc && !o.RenderBCC {
			continue
		}
		if h.Key == HdrReturnPath && !o.RenderReturnPath {
			continue
		}
		syn, ok := HeaderSyntax[h.Key]
		if ok {
			if syn.Unique {
				_, ok = seen[h.Key]
				if ok {
					continue
				}
				seen[h.Key] = struct{}{}
			}
			err := writeHeader(w, syn.Type, h.Key, h.Value, o)
			if err != nil {
				return fmt.Errorf("%s: %w", h.Key, err)
			}
			continue
		}
		err := writeHeader(w, HeaderTypeOpaque, h.Key, h.Value, o)
		if err != nil {
			return fmt.Errorf("%s: %w", h.Key, err)
		}
	}
	return nil
}

func (h *Header) Bytes(o Options) ([]byte, error) {
	var buff bytes.Buffer
	err := h.WriteTo(&buff, o)
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func Check(name, value string) error {
	headerType, ok := HeaderSyntax[textproto.CanonicalMIMEHeaderKey(name)]
	if !ok {
		return nil
	}
	return checkHeader(headerType.Type, value)
}

func checkHeader(headerType HeaderType, value string) error {
	value = strings.TrimSpace(value)
	switch headerType {
	case HeaderTypeUnstructured, HeaderTypePhraseList:
		return nil
	case HeaderTypeOpaque, HeaderTypeReceived:
		if isAscii(value) {
			return nil
		}
		return errors.New("cannot contain non-ascii characters")
	case HeaderTypeReturnPath:
		if value == "<>" {
			return nil
		}
		addr, err := mail.ParseAddress(value)
		if err != nil {
			return fmt.Errorf("'%s' is not a valid return path: %w", value, err)
		}
		if addr.Name != "" {
			return fmt.Errorf("'%s' is not a valid return path: cannot ahve display name", value)
		}
		return nil
	case HeaderTypeDate:
		return validDate(value)
	case HeaderTypeMailbox:
		_, err := mail.ParseAddress(value)
		if err == nil {
			return nil
		}
		return fmt.Errorf("'%s' is not a valid 5322 email address: %w", value, err)
	case HeaderTypeMailboxList:
		_, err := mail.ParseAddressList(value)
		if err == nil {
			return nil
		}
		return fmt.Errorf("'%s' is not a valid 5322 list of email addresses: %w", value, err)
	case HeaderTypeMessageID:
		return validMessageId(value)
	case HeaderTypeMessageIDList:
		return validMessageIdList(value)
	default:
		return fmt.Errorf("internal error, invalid header type: %v", headerType)
	}
}

// isAscii checks whether all characters in a string are low ASCII
func isAscii(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

const atext = "[a-zA-Z0-9!#$%&'*+-/=?^_`{|}~]"

func validDate(s string) error {
	_, err := mail.ParseDate(s)
	if err == nil {
		return err
	}
	return fmt.Errorf("'%s' is not a valid date: %w", s, err)
}

var messageIdRe = regexp.MustCompile(`^\s*<` + atext + `+(?:\.` + atext + `+)*@` + atext + `+(?:\.` + atext + `+)>\s*`)

func validMessageId(s string) error {
	if messageIdRe.MatchString(s) {
		return nil
	}
	return fmt.Errorf("'%s' is not a valid Message-ID", s)
}

func validMessageIdList(s string) error {
	ids := strings.Split(s, ",")
	for _, id := range ids {
		err := validMessageId(id)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeHeader(w io.Writer, headerType HeaderType, key, value string, o Options) error {
	value = strings.TrimSpace(value)
	column := len(key) + 2
	if _, err := io.WriteString(w, key); err != nil {
		return err
	}
	if _, err := io.WriteString(w, ": "); err != nil {
		return err
	}
	switch headerType {
	case HeaderTypeUnstructured, HeaderTypePhraseList:
		if !isAscii(value) && !o.NoEscape {
			value = mime.QEncoding.Encode(utf8, value)
		}
	case HeaderTypeOpaque, HeaderTypeReceived, HeaderTypeReturnPath, HeaderTypeDate, HeaderTypeMessageID, HeaderTypeMessageIDList:
	// do nothing
	case HeaderTypeMailbox:
		// TODO(steve): implement non-escaped version
		addr, err := mail.ParseAddress(value)
		if err != nil {
			return err
		}
		value = addr.String()
	case HeaderTypeMailboxList:
		// TODO(steve): implement non-escaped version
		addrs, err := mail.ParseAddressList(value)
		if err != nil {
			return err
		}
		addresses := make([]string, len(addrs))
		for i, v := range addrs {
			addresses[i] = v.String()
		}
		value = strings.Join(addresses, ", ")
	default:
		return fmt.Errorf("internal error, invalid header type: %v", headerType)
	}
	if len(value)+column < 78 {
		// simple case
		_, err := io.WriteString(w, value)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, "\r\n")
		if err != nil {
			return err
		}
		return nil
	}
	inString := false
	tokenStart := 0
	val := []byte(value)
	for i := 0; i < len(val); i++ {
		v := val[i]
		if v == '"' {
			inString = !inString
			continue
		}
		if inString {
			if v == '\r' || v == '\n' {
				return fmt.Errorf("CR or LF found in quoted string at offset %d", i)
			}
			continue
		}
		if v == '\r' || v == '\n' {
			tok := val[tokenStart:i]
			for ; i < len(val) && (val[i] == '\r' || val[i] == '\n'); i++ {
			}
			tokenStart = i
			if len(tok) > 0 {
				_, err := w.Write(tok)
				column += len(tok)
				if err != nil {
					return err
				}

				if i >= len(val) {
					break
				}
				switch val[i] {
				case ' ', '\t':
					// If user provided value includes whitespace, use that instead of a tab
					_, err = w.Write([]byte{'\r', '\n'})
					column = 0
				default:
					// Pad the continued line with a tab
					_, err = w.Write([]byte{'\r', '\n', '\t'})
					column = 1
				}
				if err != nil {
					return err
				}
			}
		}
		if v == ' ' || v == '\t' || v == '\v' || v == '\f' {
			tok := val[tokenStart:i]
			if column+len(tok) > 78 && tokenStart != 0 {
				_, err := w.Write([]byte{'\r', '\n'})
				if err != nil {
					return err
				}
				column = 0
			}
			tokenStart = i
			_, err := w.Write(tok)
			if err != nil {
				return err
			}
			column += len(tok)
		}
	}
	if tokenStart < len(val) {
		tok := val[tokenStart:]
		if column+len(tok) > 78 && tokenStart != 0 && column > 1 {
			_, err := w.Write([]byte{'\r', '\n'})
			if err != nil {
				return err
			}
			column = 0
		}
		_, err := w.Write(tok)
		if err != nil {
			return err
		}
		column += len(tok)
	}
	if column != 0 {
		_, err := w.Write([]byte{'\r', '\n'})
		if err != nil {
			return err
		}
	}
	return nil
}
