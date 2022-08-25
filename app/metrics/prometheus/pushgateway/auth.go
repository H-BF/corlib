package pushgateway

import (
	"context"
)

//Auth base interface
type Auth interface {
	isAuthPrivate()
}

//AuthProvider auth provider
type AuthProvider func(ctx context.Context) (Auth, error)

//NullAuth no auth at all
type NullAuth struct {
	Auth `json:"-"`
}

//HiddenString hidden on print
type HiddenString string

//UserBasicAuth user+password
type UserBasicAuth struct {
	Auth     `json:"-"`
	Username HiddenString
	Password HiddenString
}

//String Stringer
func (HiddenString) String() string {
	return "******"
}

//MarshalJSON json.Marshal-er
func (HiddenString) MarshalJSON() ([]byte, error) {
	return []byte("******"), nil
}
