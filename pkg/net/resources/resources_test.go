package resources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate_NetworkTransport(t *testing.T) {
	cases := []struct {
		x    NetworkTransport
		fail bool
	}{
		{TCP, false},
		{UDP, false},
		{NetworkTransport(100), true},
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

func TestValidate_Traffic(t *testing.T) {
	cases := []struct {
		x    Traffic
		fail bool
	}{
		{INGRESS, false},
		{EGRESS, false},
		{Traffic(100), true},
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

func TestValidate_FQDN(t *testing.T) {
	cases := []struct {
		val  string
		fail bool
	}{
		{"", true},
		{" ", true},
		{"*", true},
		{"*ex", false},
		{"*ex.", true},
		{"*ex.com", false},
		{"*ex.com.2", false},
		{"*ex.com.2w", false},
		{"microsoft.com", false},
	}
	for i := range cases {
		c := cases[i]
		e := FQDN(c.val).Validate()
		if !c.fail {
			require.NoErrorf(t, e, "test case #%v  '%v'", i, c.val)
		} else {
			require.Errorf(t, e, "test case #%v  '%v'", i, c.val)
		}
	}
}

func Test_Validate_ICMP(t *testing.T) {
	var x ICMP
	require.Error(t, x.Validate())
	x.IPv = 1
	require.Error(t, x.Validate())
	x.IPv = 4
	x.Types.Put(1)
	require.NoError(t, x.Validate())
}

func Test_PortSourceValid(t *testing.T) {
	require.True(t, PortSource("   ").IsValid())
	require.True(t, PortSource("  ,  ").IsValid())
	require.True(t, PortSource("  12 ").IsValid())
	require.True(t, PortSource("  12, 10, ").IsValid())
	require.True(t, PortSource("  12 - 13 ").IsValid())
	require.False(t, PortSource("   - 13 ").IsValid())
	require.True(t, PortSource("   ").IsValid())
	require.False(t, PortSource(" 12  -  ").IsValid())
	require.True(t, PortSource("").IsValid())
	require.False(t, PortSource(" a ").IsValid())
	require.False(t, PortSource(" a 10 ").IsValid())
	require.False(t, PortSource("  10 -- 13 ").IsValid())
}

func Test_PortSource2PortRange(t *testing.T) {
	eq := func(a, b PortRange) bool {
		if a == nil && b == nil {
			return true
		}
		if (a == nil && b != nil) || (a != nil && b == nil) {
			return false
		}
		l0, r0 := a.Bounds()
		l1, r1 := a.Bounds()
		return l0.Cmp(l1)|r0.Cmp(r1) == 0
	}
	cases := []struct {
		s    string
		exp  PortRange
		fail bool
	}{
		{"", nil, false},
		{" ", nil, false},
		{" 10 ", PortRangeFactory.Range(10, false, 10, false), false},
		{" 10 - 10 ", PortRangeFactory.Range(10, false, 10, false), false},
		{" 10 - 11 ", PortRangeFactory.Range(10, false, 11, false), false},
		{" - 10 - 11 ", nil, true},
		{" 11 - 10  ", nil, true},
		{" 11 - 65536  ", nil, true},
	}
	for i := range cases {
		c := cases[i]
		r, e := PortSource(c.s).ToPortRange()
		if c.fail {
			require.Errorf(t, e, "%v# '%s'", i, c.s)
		} else {
			require.NoErrorf(t, e, "%v# '%s'", i, c.s)
			require.Truef(t, eq(r, c.exp), "%v# '%s'", i, c.s)
		}
	}
}

func Test_PortSourceEq(t *testing.T) {
	cases := []struct {
		S1, S2 PortSource
		expEq  bool
	}{
		{"", " ", true},
		{"", ", , ,   ", true},
		{"10", ", , ,   ", false},
		{"10", "10", true},
		{"11, 12, 10-20", "10-20, 11", true},
		{"11, 22, 10-20", "10-20, 11", false},
	}
	for i := range cases {
		c := cases[i]
		val := c.S1.IsEq(c.S2)
		require.Equalf(t, c.expEq, val, "%v)  '%s' .EQ. '%s'", i, c.S1, c.S2)
	}
}
