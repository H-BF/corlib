package resources

import (
	"regexp"
	"strings"
	"unsafe"

	"github.com/pkg/errors"
)

// FQDN represents domain name in DNS terms
type FQDN string

// ErrInvalidFQDN -
var ErrInvalidFQDN = errors.New("invalid FQDN")

const fqdnMaxLen = 255

// String impl Stringer
func (o FQDN) String() string {
	return string(o)
}

// IsEq chacke if is Eq with no case
func (o FQDN) IsEq(other FQDN) bool {
	return strings.EqualFold(string(o), string(other))
}

// Cmp compare no case
func (o FQDN) Cmp(other FQDN) int {
	if strings.EqualFold(string(o), string(other)) {
		return 0
	}
	if o < other {
		return -1
	}
	return 1
}

// Validate impl Validator interface
func (o FQDN) Validate() error {
	a := unsafe.Slice(
		unsafe.StringData(string(o)), len(o),
	)
	if m := reFQDN.Match(a); !m || len(a) > fqdnMaxLen {
		return errors.WithMessagef(ErrInvalidFQDN, "Value('%s')", o)
	}
	return nil
}

var (
	reFQDN = regexp.MustCompile(`(?ims)^([a-z0-9\*][a-z0-9_-]{1,62}){1}(\.[a-z0-9_][a-z0-9_-]{0,62})*$`)
)
