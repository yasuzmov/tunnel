package tunnel

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	header, err := HeaderFrom(HTTPReqHeader{
		Method: "GET",
		Host:   "app.example.com",
		Path:   "/api",
		Query:  "id=1",
		Headers: map[string][]string{
			"User-Agent": {"curl/8.0"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	in := Frame{
		Type:     TypeHTTPReq,
		StreamID: 42,
		Header:   header,
		Payload:  []byte("hello"),
	}
	raw, err := EncodeFrame(in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeFrame(raw)
	if err != nil {
		t.Fatal(err)
	}
	if out.Type != in.Type || out.StreamID != in.StreamID {
		t.Fatalf("type/id mismatch: %+v", out)
	}
	if !bytes.Equal(out.Payload, in.Payload) {
		t.Fatalf("payload mismatch")
	}
	var h HTTPReqHeader
	if err := UnmarshalHeader(out, &h); err != nil {
		t.Fatal(err)
	}
	if h.Host != "app.example.com" || h.Method != "GET" {
		t.Fatalf("header mismatch: %+v", h)
	}
}

func TestDecodeShortFrame(t *testing.T) {
	if _, err := DecodeFrame([]byte{1, 2, 3}); err == nil {
		t.Fatal("expected error")
	}
}

func TestEncodeEmptyHeader(t *testing.T) {
	raw, err := EncodeFrame(Frame{Type: TypePing, StreamID: 0})
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeFrame(raw)
	if err != nil {
		t.Fatal(err)
	}
	if string(out.Header) != "{}" {
		t.Fatalf("want empty object, got %s", out.Header)
	}
}
