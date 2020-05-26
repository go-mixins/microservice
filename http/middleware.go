package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/go-mixins/log"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/xerrors"

	"github.com/go-noodle/bind"
	"github.com/go-noodle/noodle"
	"github.com/go-noodle/render"
)

var marshaler = &jsonpb.Marshaler{
	Indent:       "\t",
	EmitDefaults: true,
	OrigName:     true,
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
			return marshaler.Marshal(w, pb)
		}
		return json.NewEncoder(w).Encode(dest)
	}, "application/json")
}

var unmarshaler = &jsonpb.Unmarshaler{
	AllowUnknownFields: true,
}

type decoder struct {
	r io.Reader
}

func (d decoder) Decode(dest interface{}) (rErr error) {
	if pb, ok := dest.(proto.Message); ok {
		return xerrors.Errorf("decode jsonpb: %w", unmarshaler.Unmarshal(d.r, pb))
	}
	return xerrors.Errorf("decode json: %w", json.NewDecoder(d.r).Decode(dest))
}

func jsonPB(r *http.Request) bind.Decoder {
	return decoder{r.Body}
}

// JSON returns middleware constructor that allows binding of
// proto.Messages and generic objects from request body
func JSON(model interface{}, opts ...bind.Option) noodle.Middleware {
	return bind.Generic(model, jsonPB, opts...)
}
