package orderedheaders

import (
	"bufio"
	"net/textproto"
	"reflect"
	"strings"
	"testing"
)

func reader(s string) *textproto.Reader {
	return textproto.NewReader(bufio.NewReader(strings.NewReader(s)))
}

func TestReadMIMEHeader(t *testing.T) {
	r := reader("my-key: Value 1  \r\nLong-key: Even \n Longer Value\r\nmy-Key: Value 2\r\n\n")
	m, err := ReadHeader(r)
	want := Header{
		Headers: []KV{
			KV{Key: "My-Key", Value: "Value 1"},
			KV{Key: "Long-Key", Value: "Even Longer Value"},
			KV{Key: "My-Key", Value: "Value 2"},
		},
	}

	if !reflect.DeepEqual(m, want) || err != nil {
		t.Fatalf("ReadMIMEHeader: %v, %v; want %v", m, err, want)
	}

	wantMap := textproto.MIMEHeader{
		"My-Key":   {"Value 1", "Value 2"},
		"Long-Key": {"Even Longer Value"},
	}

	tpm := m.ToMap()
	if !reflect.DeepEqual(tpm, wantMap) || err != nil {
		t.Fatalf("ReadMIMEHeader: %v, %v; want %v", tpm, err, wantMap)
	}
}

func TestReadMIMEHeaderSingle(t *testing.T) {
	r := reader("Foo: bar\n\n")
	m, err := ReadHeader(r)
	want := Header{
		Headers: []KV{
			KV{Key: "Foo", Value: "bar"},
		},
	}

	if !reflect.DeepEqual(m, want) || err != nil {
		t.Fatalf("ReadMIMEHeader: %v, %v; want %v", m, err, want)
	}

	wantMap := textproto.MIMEHeader{"Foo": {"bar"}}

	tpm := m.ToMap()
	if !reflect.DeepEqual(tpm, wantMap) || err != nil {
		t.Fatalf("ReadMIMEHeader: %v, %v; want %v", tpm, err, wantMap)
	}
}

func TestReadMIMEHeaderNoKey(t *testing.T) {
	r := reader(": bar\ntest-1: 1\n\n")
	m, err := ReadHeader(r)
	want := Header{
		Headers: []KV{
			KV{Key: "Test-1", Value: "1"},
		},
	}

	if !reflect.DeepEqual(m, want) || err != nil {
		t.Fatalf("ReadMIMEHeader: %v, %v; want %v", m, err, want)
	}

	wantMap := textproto.MIMEHeader{"Test-1": {"1"}}

	tpm := m.ToMap()
	if !reflect.DeepEqual(tpm, wantMap) || err != nil {
		t.Fatalf("ReadMIMEHeader: %v, %v; want %v", tpm, err, wantMap)
	}
}

func TestLargeReadMIMEHeader(t *testing.T) {
	data := make([]byte, 16*1024)
	for i := 0; i < len(data); i++ {
		data[i] = 'x'
	}
	sdata := string(data)
	r := reader("Cookie: " + sdata + "\r\n\n")
	m, err := ReadHeader(r)
	if err != nil {
		t.Fatalf("ReadMIMEHeader: %v", err)
	}
	cookie := m.Get("Cookie")
	if cookie != sdata {
		t.Fatalf("ReadMIMEHeader: %v bytes, want %v bytes", len(cookie), len(sdata))
	}
}

// Test that we read slightly-bogus MIME headers seen in the wild,
// with spaces before colons, and spaces in keys.
func TestReadMIMEHeaderNonCompliant(t *testing.T) {
	// Invalid HTTP response header as sent by an Axis security
	// camera: (this is handled by IE, Firefox, Chrome, curl, etc.)
	r := reader("Foo: bar\r\n" +
		"Content-Language: en\r\n" +
		"SID : 0\r\n" +
		"Audio Mode : None\r\n" +
		"Privilege : 127\r\n\r\n")
	m, err := ReadHeader(r)
	want := Header{
		Headers: []KV{
			KV{Key: "Foo", Value: "bar"},
			KV{Key: "Content-Language", Value: "en"},
			KV{Key: "Sid", Value: "0"},
			KV{Key: "Audio Mode", Value: "None"},
			KV{Key: "Privilege", Value: "127"},
		},
	}

	if !reflect.DeepEqual(m, want) || err != nil {
		t.Fatalf("ReadMIMEHeader =\n%v, %v; want:\n%v", m, err, want)
	}

	wantMap := textproto.MIMEHeader{
		"Foo":              {"bar"},
		"Content-Language": {"en"},
		"Sid":              {"0"},
		"Audio Mode":       {"None"},
		"Privilege":        {"127"},
	}

	tpm := m.ToMap()
	if !reflect.DeepEqual(tpm, wantMap) || err != nil {
		t.Fatalf("ReadMIMEHeader: %v, %v; want %v", tpm, err, wantMap)
	}
}

// Test that continued lines are properly trimmed.
func TestReadMIMEHeaderTrimContinued(t *testing.T) {
	// In this header, \n and \r\n terminated lines are mixed on purpose.
	// We expect each line to be trimmed (prefix and suffix) before being concatenated.
	// Keep the spaces as they are.
	r := reader("" + // for code formatting purpose.
		"a:\n" +
		" 0 \r\n" +
		"b:1 \t\r\n" +
		"c: 2\r\n" +
		" 3\t\n" +
		"  \t 4  \r\n\n")
	m, err := ReadHeader(r)
	if err != nil {
		t.Fatal(err)
	}
	want := Header{
		Headers: []KV{
			KV{Key: "A", Value: "0"},
			KV{Key: "B", Value: "1"},
			KV{Key: "C", Value: "2 3 4"},
		},
	}

	if !reflect.DeepEqual(m, want) {
		t.Fatalf("ReadMIMEHeader mismatch.\n got: %q\nwant: %q", m, want)
	}

	wantMap := textproto.MIMEHeader{
		"A": {"0"},
		"B": {"1"},
		"C": {"2 3 4"},
	}

	tpm := m.ToMap()
	if !reflect.DeepEqual(tpm, wantMap) || err != nil {
		t.Fatalf("ReadMIMEHeader: %v, %v; want %v", tpm, err, wantMap)
	}
}
