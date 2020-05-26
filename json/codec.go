package json

import (
	"bytes"
	"io"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"

	"golang.org/x/xerrors"
)

// Дефолтные кодеки
var (
	DefaultMarshaler = &jsonpb.Marshaler{
		Indent:       "\t",
		EmitDefaults: true,
		OrigName:     true,
	}

	DefaultUnmarshaler = &jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}
)

// Unmarshal convenience wrapper
func Unmarshal(r io.Reader, pb proto.Message) error {
	if err := DefaultUnmarshaler.Unmarshal(r, pb); err != nil {
		return xerrors.Errorf("unmarshal JSON: %w")
	}
	return nil
}

// Marshal convenience wrapper
func Marshal(out io.Writer, pb proto.Message) error {
	if err := DefaultMarshaler.Marshal(out, pb); err != nil {
		return xerrors.Errorf("marshal to JSON: %w", err)
	}
	return nil
}

// Encode object to JSON
func Encode(src proto.Message) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := DefaultMarshaler.Marshal(buf, src); err != nil {
		return nil, xerrors.Errorf("encode JSON: %w", err)
	}
	return buf.Bytes(), nil
}

// Decode object from JSON
func Decode(data []byte, dest proto.Message) error {
	if err := DefaultUnmarshaler.Unmarshal(bytes.NewReader(data), dest); err != nil {
		return xerrors.Errorf("decode JSON: %w", err)
	}
	return nil
}
