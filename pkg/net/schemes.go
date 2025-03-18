package net

import (
	"net/url"
	"strings"
)

// Schema ...
type Schema string

// Is check is s == another
func (s Schema) Is(another any) bool {
	var s1 string
	switch v := another.(type) {
	case string:
		s1 = v
	case []byte:
		s1 = string(v)
	case Schema:
		s1 = string(v)
	case Endpoint:
		s1 = v.Network()
	case *Endpoint:
		s1 = v.Network()
	case url.URL:
		s1 = v.Scheme
	case *url.URL:
		s1 = v.Scheme
	default:
		return false
	}
	return strings.EqualFold(string(s), s1)
}

type shemaHolderIFace interface {
	Endpoint | ~*Endpoint | ~string | ~[]byte | url.URL | ~*url.URL
}

// AnySchema -
func AnySchema[tArg shemaHolderIFace](schemaHolder tArg, sh ...Schema) bool {
	for _, s := range sh {
		if s.Is(schemaHolder) {
			return true
		}
	}
	return false
}

const (
	//SchemeUnixHTTP обозначим схему как http-unix://[/]socket_path/path[?query]
	SchemeUnixHTTP Schema = "http+unix"
	//SchemeHTTPS like https://
	SchemeHTTPS Schema = "https"
	//SchemeHTTP like http://
	SchemeHTTP Schema = "http"
)

const (
	// SchemeTCP like tcp:[//]
	SchemeTCP Schema = "tcp"
	// SchemeUDP like udp:[//]
	SchemeUDP Schema = "udp"
	// SchemeUNIX like unix:[//]
	SchemeUNIX Schema = "unix"
)
