package jsonview

import (
	"bytes"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type (
	stringer struct {
		d interface{}
	}

	marshaler func() ([]byte, error)
)

//String ...
func (s *stringer) String() string {
	return String(s.d)
}

//MarshalJSON ...
func (f marshaler) MarshalJSON() ([]byte, error) {
	return f()
}

//String ...
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

//Stringer ...
func Stringer(d interface{}) fmt.Stringer {
	return &stringer{d: d}
}

//Marshaler ...
func Marshaler(d interface{}) json.Marshaler {
	switch t := d.(type) {
	case nil:
		return nil
	case json.Marshaler:
		return t
	case error:
		return marshaler(func() ([]byte, error) {
			b := bytes.NewBuffer(nil)
			_, e := fmt.Fprintf(b, "%q", t)
			return b.Bytes(), e
		})
	case proto.Message:
		return marshaler(func() ([]byte, error) {
			return protojson.MarshalOptions{AllowPartial: true}.Marshal(t)
		})
	case fmt.Stringer:
		return marshaler(func() ([]byte, error) {
			b := bytes.NewBuffer(nil)
			_, e := fmt.Fprintf(b, "%q", t)
			return b.Bytes(), e
		})
	case fmt.GoStringer:
		return marshaler(func() ([]byte, error) {
			b := bytes.NewBuffer(nil)
			_, e := fmt.Fprintf(b, "%q", fmt.Sprintf("%#v", t))
			return b.Bytes(), e
		})
	}
	return marshaler(func() ([]byte, error) {
		return json.Marshal(d)
	})
}
