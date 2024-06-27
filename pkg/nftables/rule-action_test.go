//go:build linux
// +build linux

package nftables

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate_RuleAction(t *testing.T) {
	cases := []struct {
		x    RuleAction
		fail bool
	}{
		{RA_DROP, false},
		{RA_ACCEPT, false},
		{RuleAction(100), true},
	}
	for i := range cases {
		c := cases[i]
		e := c.x.Validate()
		if !c.fail {
			require.NoErrorf(t, e, "test case #%v", i)
		} else {
			require.Errorf(t, e, "test case #%v", i)
		}
	}
}
