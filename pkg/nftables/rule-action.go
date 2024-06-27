//go:build linux
// +build linux

package nftables

import (
	"fmt"
	"strings"

	oz "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// RuleAction terminal verdict action for rules
type RuleAction uint8

const (
	// RA_UNDEF -
	RA_UNDEF RuleAction = iota

	// RA_DROP setups rule to drop packet
	RA_DROP

	// RA_ACCEPT setups rule to accept packet
	RA_ACCEPT
)

// String impl Stringer
func (a RuleAction) String() string {
	return [...]string{"undef", "drop", "accept"}[a]
}

// IsEq -
func (a RuleAction) IsEq(other RuleAction) bool {
	return a == other
}

// FromString init from string
func (a *RuleAction) FromString(s string) error {
	const api = "RuleAction/FromString"
	switch strings.ToLower(s) {
	case "drop":
		*a = RA_DROP
	case "accept":
		*a = RA_ACCEPT
	default:
		return errors.WithMessage(fmt.Errorf("unknown value '%s'", s), api)
	}
	return nil
}

// Validate RuleAction validator
func (a RuleAction) Validate() error {
	vals, x := [...]any{int(RA_DROP), int(RA_ACCEPT)}, int(a)
	return oz.Validate(x, oz.In(vals[:]...).Error("must be in ['DROP', 'ACCEPT']"))
}
