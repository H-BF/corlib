package resources

import (
	"bytes"
	"fmt"

	"github.com/H-BF/corlib/pkg/dict"

	oz "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// IPv4 IP family v4
	IPv4 uint8 = 4
	// IPv6 IP family v6
	IPv6 uint8 = 6
)

// ICMP an ICMP proto spec
type ICMP struct {
	IPv   uint8             // Use in IP net version 4 or 6
	Types dict.RBSet[uint8] // Use ICMP message types set of [0-254]
}

// IsEq -
func (o ICMP) IsEq(other ICMP) bool {
	return o.IPv == other.IPv &&
		o.Types.Eq(&other.Types)
}

// Validate -
func (o ICMP) Validate() error {
	return oz.ValidateStruct(&o,
		oz.Field(&o.IPv, oz.Required, oz.In(uint8(IPv4), uint8(IPv6)).
			Error("IPv should be in [4,6]")),
	)
}

// String -
func (o ICMP) String() string {
	b := bytes.NewBuffer(nil)
	_, _ = b.WriteString("ICMP")
	if o.IPv == IPv6 {
		_ = b.WriteByte('6')
	}
	if i := 0; o.Types.Len() > 0 {
		o.Types.Iterate(func(k uint8) bool {
			if i++; i > 1 {
				_ = b.WriteByte(',')
			}
			_, _ = fmt.Fprintf(b, "%v", k)
			return true
		})
	}
	return b.String()
}
