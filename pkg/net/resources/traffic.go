package resources

import (
	"fmt"
	"strings"

	oz "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

// Traffic packet traffic any of [INGRESS, EGRESS]
type Traffic uint8

const (
	// INGRESS as is
	INGRESS Traffic = iota + 1

	// EGRESS as is
	EGRESS
)

// String -
func (tfc Traffic) String() string {
	switch tfc {
	case INGRESS:
		return "ingress"
	case EGRESS:
		return "egress"
	}
	return fmt.Sprintf("Undef(%v)", int(tfc))
}

// FromString init from string
func (tfc *Traffic) FromString(s string) error {
	const (
		ing = "ingress"
		egr = "egress"
	)
	const api = "Traffic/FromString"
	switch strings.ToLower(s) {
	case ing:
		*tfc = INGRESS
	case egr:
		*tfc = EGRESS
	default:
		return errors.WithMessage(fmt.Errorf("unknown value '%s'", s), api)
	}
	return nil
}

// IsEq -
func (tfc Traffic) IsEq(other Traffic) bool {
	return tfc == other
}

// Validate net transport validator
func (tfc Traffic) Validate() error {
	vals, x := [...]any{int(INGRESS), int(EGRESS)}, int(tfc)
	return oz.Validate(x, oz.In(vals[:]...).Error("must be in ['INGRESS', 'EGRESS']"))
}
