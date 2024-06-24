package resources

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type portSourceHelper struct{}

func (h portSourceHelper) str2portranges(ps string, sep string) ([]PortRange, error) {
	if sep == "" {
		panic("invalid separator")
	}
	var ret []PortRange
	for _, s := range strings.Split(ps, sep) {
		r, e := h.str2portrange(s)
		if e != nil {
			return nil, e
		}
		if r != nil && !r.IsNull() {
			ret = append(ret, r)
		}
	}
	return ret, nil
}

func (portSourceHelper) str2portrange(ps string) (PortRange, error) {
	var (
		err  error
		l, r uint64
	)
	m := parsePortsRE.FindStringSubmatch(ps)
	if len(m) != 4 { //nolint:mnd
		return nil,
			errors.WithMessagef(errIncorrectPortsSource, "unrecognized value '%s'", ps)
	}
	if m[2] != "" && m[3] != "" {
		l, err = strconv.ParseUint(m[2], 10, 16)
		if err == nil {
			r, err = strconv.ParseUint(m[3], 10, 16)
			if err == nil && l > r {
				return nil, errors.WithMessagef(errIncorrectPortsSource,
					"the left bound '%v' is greather than right one '%v'", l, r)
			}
		}
	} else if m[1] != "" {
		l, err = strconv.ParseUint(m[1], 10, 16)
		r = l
	} else {
		return nil, nil
	}
	if err != nil {
		return nil, multierr.Combine(errIncorrectPortsSource, err)
	}
	return PortRangeFactory.Range(
		PortNumber(l), false,
		PortNumber(r), false,
	), nil
}

func (portSourceHelper) fromPortRange(r PortRange, w io.Writer) error {
	if r == nil || r.IsNull() {
		return nil
	}
	lb, rb := r.Bounds()
	if _, excl := lb.GetValue(); excl {
		lb = rb.AsIncluded()
		if _, excl = lb.GetValue(); excl {
			return errIncorrectPortsSource
		}
	}
	if _, excl := rb.GetValue(); excl {
		rb = rb.AsIncluded()
		if _, excl = rb.GetValue(); excl {
			return errIncorrectPortsSource
		}
	}
	vr, _ := rb.GetValue()
	vl, _ := lb.GetValue()
	if vl == vr {
		fmt.Fprintf(w, "%v", vl)
	} else {
		fmt.Fprintf(w, "%v-%v", vl, vr)
	}
	return nil
}

var (
	errIncorrectPortsSource = fmt.Errorf("incorrect port range(s) source")
	parsePortsRE            = regexp.MustCompile(`^\s*((?:(\d+)\s*-\s*(\d+))|\d+|\s*)\s*$`)
)
