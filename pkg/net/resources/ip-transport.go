package resources

import (
	"fmt"
	"strings"

	oz "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

const (
	// TCP -
	TCP NetworkTransport = iota

	// UDP -
	UDP
)

// NetworkTransport net transport
type NetworkTransport uint8

// String impl Stringer
func (nt NetworkTransport) String() string {
	return [...]string{"tcp", "udp"}[nt]
}

// FromString init from string
func (nt *NetworkTransport) FromString(s string) error {
	const api = "NetworkTransport/FromString"
	switch strings.ToLower(s) {
	case "tcp":
		*nt = TCP
	case "udp":
		*nt = UDP
	default:
		return errors.WithMessage(fmt.Errorf("unknown value '%s'", s), api)
	}
	return nil
}

// IsEq -
func (nt NetworkTransport) IsEq(other NetworkTransport) bool {
	return nt == other
}

// Validate net transport validator
func (nt NetworkTransport) Validate() error {
	vals, x := [...]any{int(TCP), int(UDP)}, int(nt)
	return oz.Validate(x, oz.In(vals[:]...).Error("must be in ['TCP', 'UDP']"))
}
