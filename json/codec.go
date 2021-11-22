package json

import (
	"fmt"
	"io"
	"io/ioutil"

	jsonpb "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Дефолтные кодеки
var (
	DefaultMarshaler = &jsonpb.MarshalOptions{
		Multiline:       true,
		Indent:          "\t",
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}

	DefaultUnmarshaler = &jsonpb.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

// Unmarshal convenience wrapper
func Unmarshal(r io.Reader, pb proto.Message) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	if err := DefaultUnmarshaler.Unmarshal(data, pb); err != nil {
		return fmt.Errorf("unmarshal JSON: %w", err)
	}
	return nil
}

// Marshal convenience wrapper
func Marshal(out io.Writer, pb proto.Message) error {
	data, err := DefaultMarshaler.Marshal(pb)
	if err != nil {
		return fmt.Errorf("marshal to JSON: %w", err)
	}
	_, err = out.Write(data)
	return err
}

// Encode object to JSON
func Encode(src proto.Message) ([]byte, error) {
	return DefaultMarshaler.Marshal(src)
}

// Decode object from JSON
func Decode(data []byte, dest proto.Message) error {
	return DefaultUnmarshaler.Unmarshal(data, dest)
}
