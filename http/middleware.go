package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/go-mixins/log"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/charmap"
	jsonpb "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/go-noodle/bind"
	"github.com/go-noodle/noodle"
	"github.com/go-noodle/render"
)

var marshaler = &jsonpb.MarshalOptions{
	Multiline:       true,
	Indent:          "\t",
	EmitUnpopulated: true,
	UseProtoNames:   true,
}

// Debug incoming HTTP requests
func Debug() noodle.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			data, _ := httputil.DumpRequest(r, true)
			if enc, _, ok := charset.DetermineEncoding(data, r.Header.Get("Content-Type")); ok {
				if decoded, err := enc.NewDecoder().Bytes(data); err == nil {
					data = decoded
				}
			} else if bytes.Contains(data, []byte("windows-1251")) {
				data, _ = charmap.Windows1251.NewDecoder().Bytes(data)
			}
			if len(data) > 4096 {
				data = append(data[:4096], []byte("...")...)
			}
			log.Get(r.Context()).Debugf("received request: %s", data)
			next(w, r)
		}
	}
}

// Render extends noodle's render middleware with JSONPB support
func Render() noodle.Middleware {
	return render.Generic(func(w io.Writer, dest interface{}) error {
		if dest == nil {
			return nil
		}
		if pb, ok := dest.(proto.Message); ok {
			res, err := marshaler.Marshal(pb)
			if err != nil {
				return err
			}
			if _, err := w.Write(res); err != nil {
				return err
			}
		}
		return json.NewEncoder(w).Encode(dest)
	}, "application/json")
}

var unmarshaler = &jsonpb.UnmarshalOptions{
	DiscardUnknown: true,
}

type decoder struct {
	r io.Reader
}

func (d decoder) Decode(dest interface{}) (rErr error) {
	if pb, ok := dest.(proto.Message); ok {
		data, err := ioutil.ReadAll(d.r)
		if err != nil {
			return err
		}
		return fmt.Errorf("decode jsonpb: %w", unmarshaler.Unmarshal(data, pb))
	}
	return fmt.Errorf("decode json: %w", json.NewDecoder(d.r).Decode(dest))
}

func jsonPB(r *http.Request) bind.Decoder {
	return decoder{r.Body}
}

// JSON returns middleware constructor that allows binding of
// proto.Messages and generic objects from request body
func JSON(model interface{}, opts ...bind.Option) noodle.Middleware {
	return bind.Generic(model, jsonPB, opts...)
}
