package nftables

import (
	"fmt"
	"strings"

	oz "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// ChainDefaultAction default action for SG {DROP|ACCEPT}
type ChainDefaultAction uint8

const (
	// DEFAULT is mean default action
	DEFAULT ChainDefaultAction = iota

	// DROP drop action net packet
	DROP

	// ACCEPT accept action net packet
	ACCEPT
)

// String impl Stringer
func (a ChainDefaultAction) String() string {
	return [...]string{"default", "drop", "accept"}[a]
}

// FromString inits from string
func (a *ChainDefaultAction) FromString(s string) error {
	const api = "ChainDefaultAction/FromString"
	switch strings.ToLower(s) {
	case "defuault":
		*a = DEFAULT
	case "drop":
		*a = DROP
	case "accept":
		*a = ACCEPT
	default:
		return errors.WithMessage(fmt.Errorf("unknown value '%s'", s), api)
	}
	return nil
}

// Validate ChainDefaultAction validator
func (a ChainDefaultAction) Validate() error {
	vals, x := [...]any{int(DEFAULT), int(DROP), int(ACCEPT)}, int(a)
	return oz.Validate(x, oz.In(vals[:]...).Error("must be in ['DEFAULT', 'DROP', 'ACCEPT']"))
}
