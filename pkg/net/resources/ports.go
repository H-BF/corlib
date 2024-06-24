package resources

import (
	"bytes"
	"strings"

	"github.com/H-BF/corlib/pkg/ranges"
)

type (
	// PortSource -
	PortSource string

	// PortNumber net port num
	PortNumber = uint16

	// PortRanges net port ranges
	PortRanges = ranges.MultiRange[PortNumber]

	// PortRange net port range
	PortRange = ranges.Range[PortNumber]
)

// PortRangeFactory ...
var PortRangeFactory = ranges.IntsFactory(PortNumber(0))

// PortRangeFull port range [0, 65535]
var PortRangeFull = PortRangeFactory.Range(0, false, ^PortNumber(0), false)

// NewPortRarnges is a port rarnges constructor
func NewPortRarnges() PortRanges {
	return ranges.NewMultiRange(PortRangeFactory)
}

// IsValid check string of port range is valid
func (ps PortSource) IsValid() bool {
	for _, s := range strings.Split(string(ps), ",") {
		m := parsePortsRE.FindStringSubmatch(s)
		if !(len(m) != 0 && m[0] == s) {
			return false
		}
	}
	return true
}

// IsEq -
func (ps PortSource) IsEq(other PortSource) bool {
	var h portSourceHelper
	var rr1, rr2 []PortRange
	var e error
	if rr1, e = h.str2portranges(string(ps), ","); e != nil {
		return false
	}
	if rr2, e = h.str2portranges(string(other), ","); e != nil {
		return false
	}
	pr := NewPortRarnges()
	pr.Update(ranges.CombineExclude, append(rr1, rr2...)...)
	return pr.Len() == 0
}

// FromPortRange inits from PortRange
func (ps *PortSource) FromPortRange(r PortRange) error {
	buf := bytes.NewBuffer(nil)
	if e := (portSourceHelper{}).fromPortRange(r, buf); e != nil {
		return e
	}
	*ps = PortSource(buf.String())
	return nil
}

// FromPortRanges -
func (ps *PortSource) FromPortRanges(rr PortRanges) error {
	buf := bytes.NewBuffer(nil)
	var e error
	rr.Iterate(func(r PortRange) bool {
		if r.IsNull() {
			return true
		}
		if buf.Len() > 0 {
			_ = buf.WriteByte(',')
		}
		e = portSourceHelper{}.fromPortRange(r, buf)
		return e == nil
	})
	if e == nil {
		*ps = PortSource(buf.String())
	}
	return e
}

// ToPortRange string to port range
func (ps PortSource) ToPortRange() (PortRange, error) {
	ret, e := portSourceHelper{}.str2portrange(string(ps))
	if e == nil && ret != nil && ret.IsNull() {
		ret = nil
	}
	return ret, e
}

// ToPortRanges -
func (ps PortSource) ToPortRanges() (PortRanges, error) {
	ret := NewPortRarnges()
	src, err := portSourceHelper{}.str2portranges(string(ps), ",")
	if err != nil {
		return ret, err
	}
	ret.Update(ranges.CombineMerge, src...)
	return ret, nil
}
