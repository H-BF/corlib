package net

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Endpoint endpoint to connect to
type Endpoint struct {
	endpointIFaceBaseImpl
}

// ParseEndpoint parse endpoint address
func ParseEndpoint(src string) (*Endpoint, error) {
	const (
		api  = "ParseEndpoint"
		unix = "unix"
		tcp  = "tcp"
		tcp4 = "tcp4"
		tcp6 = "tcp6"
		udp  = "udp"
		udp4 = "udp4"
		udp6 = "udp6"
	)
	parts := reSchemaAndAddress.FindStringSubmatch(src)
	if len(parts) < 3 {
		return nil, errors.Errorf("%s: address '%s' seems invalid", api, src)
	}
	ep := new(Endpoint)
	schema := strings.ToLower(parts[1])
	addr := parts[2]
	switch schema {
	case unix:
		ep.endpointIFaceBaseImpl.delegate = endpointAddressUnix{socketPath: addr}
	case tcp, tcp4, tcp6, "", udp, udp4, udp6:
		h, p, e := net.SplitHostPort(addr)
		if e != nil {
			return nil, errors.Errorf("%s: the addr '%s' has invalid host:port", api, src)
		}
		h, p = strings.TrimSpace(h), strings.TrimSpace(p)
		if len(p) == 0 {
			return nil, errors.Errorf("%s: in addr('%s') port is empty", api, src)
		}
		for _, s := range []*string{&h, &p} {
			if strings.ContainsAny(*s, "/?\\=% #") {
				return nil, errors.Errorf("%s: the addr('%s') is invalid", api, src)
			}
		}
		if strings.HasPrefix(schema, string(SchemeTCP)) || schema == "" {
			ep.endpointIFaceBaseImpl.delegate = endpointAddressTCP{host: h, port: p}
		} else {
			ep.endpointIFaceBaseImpl.delegate = endpointAddressUDP{host: h, port: p}
		}
	default:
		return nil, errors.Errorf("%s: the addr '%s' has unsupported schema '%s'", api, src, schema)
	}
	return ep, nil
}

func (ep *Endpoint) String() string {
	s, _ := ep.Address()
	return s
}

// HostPort gives host - port if TCP case is
func (ep *Endpoint) HostPort() (host, port string, err error) {
	switch t := ep.endpointIFaceBaseImpl.delegate.(type) {
	case endpointAddressTCP:
		host, port = t.host, t.port
	default:
		err = errors.Errorf("endpoint is neither TCP nor UDP")
	}
	return
}

// IsUnixDomain returns true when endpoint is unix domain socket
func (ep *Endpoint) IsUnixDomain() bool {
	return ep.Network() == string(SchemeUNIX)
}

// FQN full qualified name
func (ep *Endpoint) FQN() string {
	if a, _ := ep.Address(); len(a) > 0 {
		nw := ep.Network()
		if len(nw) == 0 {
			return a
		}
		return fmt.Sprintf("%s://%s", nw, a)
	}
	return ""
}

var (
	reSchemaAndAddress          = regexp.MustCompile(`^\s*(?:(\w+):(?://)?)?([^\\\s#?=]+)\s*$`)
	_                           = ParseEndpoint
	_                  net.Addr = (*Endpoint)(nil)
)

type (
	endpointIFace interface {
		Network() string
		Address() (string, error)
	}

	endpointAddressTCP struct {
		host string
		port string
	}

	endpointAddressUDP struct {
		host string
		port string
	}

	endpointAddressUnix struct {
		socketPath string
	}

	endpointIFaceBaseImpl struct {
		delegate endpointIFace
	}
)

// Network impl endpointIFace
func (bi endpointIFaceBaseImpl) Network() string {
	if bi.delegate != nil {
		return bi.delegate.Network()
	}
	return ""
}

// Address impl endpointIFace
func (bi endpointIFaceBaseImpl) Address() (string, error) {
	if bi.delegate != nil {
		return bi.delegate.Address()
	}
	return "", errors.New("not initialized")
}

// Network impl endpointIFace
func (endpointAddressTCP) Network() string {
	return string(SchemeTCP)
}

// Address impl endpointIFace
func (t endpointAddressTCP) Address() (string, error) {
	return net.JoinHostPort(t.host, t.port), nil
}

// Network impl endpointIFace
func (endpointAddressUDP) Network() string {
	return string(SchemeUDP)
}

// Address impl endpointIFace
func (t endpointAddressUDP) Address() (string, error) {
	return net.JoinHostPort(t.host, t.port), nil
}

// Network impl endpointIFace
func (endpointAddressUnix) Network() string {
	return string(SchemeUNIX)
}

// Address impl endpointIFace
func (t endpointAddressUnix) Address() (string, error) {
	return t.socketPath, nil
}
