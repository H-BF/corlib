package jsonview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type (
	stringer struct {
		d interface{}
	}

	marshaler func() ([]byte, error)
)

// String ...
func (s *stringer) String() string {
	return String(s.d)
}

// MarshalJSON ...
func (f marshaler) MarshalJSON() ([]byte, error) {
	return f()
}

// String ...
func String(d interface{}) string {
	switch t := d.(type) {
	case nil:
		return "nil"
	case error:
		return t.Error()
	case proto.Message:
		b, _ := protojson.MarshalOptions{AllowPartial: true}.Marshal(t)
		return string(b)
	case fmt.Stringer:
		return t.String()
	case fmt.GoStringer:
		return t.GoString()
	}
	b, _ := json.Marshal(d)
	return string(b)
}

// Stringer ...
func Stringer(d interface{}) fmt.Stringer {
	return &stringer{d: d}
}

// Marshaler ...
func Marshaler(d interface{}) json.Marshaler {
	switch t := d.(type) {
	case nil:
		return nil
	case json.Marshaler:
		return t
	case error:
		return quoteMarshaler(t.Error())
	case proto.Message:
		return marshaler(func() ([]byte, error) {
			return protojson.MarshalOptions{AllowPartial: true}.Marshal(t)
		})
	case net.Addr:
		switch addr := t.(type) {
		case *net.TCPAddr:
			return quoteMarshaler(addr.IP.String())
		case *net.UnixAddr:
			return marshaler(func() ([]byte, error) {
				return []byte("\"unix-socket\""), nil
			})
		default:
			return quoteMarshaler(addr.String())
		}
	case fmt.Stringer:
		return quoteMarshaler(t.String())
	case fmt.GoStringer:
		return quoteMarshaler(fmt.Sprintf("%#v", t))
	}
	return marshaler(func() ([]byte, error) {
		return json.Marshal(d)
	})
}

func quoteMarshaler(value string) marshaler {
	return marshaler(func() ([]byte, error) {
		b := bytes.NewBuffer(nil)
		_, e := fmt.Fprintf(b, "%q", value)
		return b.Bytes(), e
	})
}
