package ranges

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"

	"github.com/pkg/errors"
)

type Ints interface {
	uint8 | uint16 | uint32 | uint64 | uint |
		int8 | int16 | int32 | int64 | int
}

type ParseSources interface {
	~string | ~[]byte
}

type intsRange[T Ints] struct {
	lower intsBound[T, Lower]
	upper intsBound[T, Upper]
}

var (
	//ParseError ...
	ParseError = errors.New("parse error")

	//SourceMatchError ...
	SourceMatchError = errors.New("source doesn't match to 'ranges'")
)

var (
	_ Range[int] = (*intsRange[int])(nil)
)

func (i *intsRange[T]) Copy() Range[T] {
	ret := *i
	return &ret
}

func (i *intsRange[T]) Normalize() Range[T] {
	ret := i.Copy()
	if !i.IsNull() {
		rL, rU := ret.Bounds()
		ch := 0
		if v, excl := rL.GetValue(); excl && v+1 > v {
			rL.SetValue(v+1, false)
			ch++
		}
		if v, excl := rU.GetValue(); !excl && v+1 > v {
			rU.SetValue(v+1, true)
			ch++
		}
		if ch > 0 {
			ret.SetBounds(rL, rU)
		}
	}
	return ret
}

func (i *intsRange[T]) Bounds() (lower, upper Bound[T]) {
	return i.lower.Copy(), i.upper.Copy()
}

func (i *intsRange[T]) SetBounds(lower, upper Bound[T]) {
	if lower.Is(Upper{}) {
		lower = lower.Adjacent()
	}
	if upper.Is(Lower{}) {
		upper = upper.Adjacent()
	}
	i.lower.SetValue(lower.GetValue())
	i.upper.SetValue(upper.GetValue())
}

func (i *intsRange[T]) String() string {
	return fmt.Sprintf("%s,%s", &i.lower, &i.upper)
}

func (i *intsRange[T]) Contains(v T) bool {
	lV, lE := i.lower.GetValue()
	uV, uE := i.upper.GetValue()
	return (lV < v && v < uV) ||
		(!lE && v == lV) ||
		(!uE && v == uV)
}

func (i *intsRange[T]) Filter(f func(x T, included bool) bool, data ...T) {
	for _, d := range data {
		if contains := i.Contains(d); !f(d, contains) {
			break
		}
	}
}

func (i *intsRange[T]) IsNull() bool {
	lV, lE := i.lower.GetValue()
	uV, uE := i.upper.GetValue()
	if lV <= uV {
		var d T
		if lE {
			d++
		}
		if uE {
			d++
		}
		return (uV - lV) < d
	}
	return true
}

func ParseIntsRange[T Ints, S ParseSources](in S, result *Range[T]) error {
	const (
		msgUnexpected = "unexpected behaviour reached"
	)

	source := reflect.ValueOf(in).Convert(
		reflect.TypeOf((*[]byte)(nil)).Elem(),
	).Interface().([]byte)

	re := intsRangeRe.FindSubmatchIndex(source)
	if len(re) == 0 {
		return SourceMatchError
	}

	b := bytes.NewBuffer(nil)
	_, _ = b.Write(source[re[2]:re[3]])
	_, _ = b.Write(source[re[4]:re[5]])
	_ = b.WriteByte(',')
	_, _ = b.Write(source[re[6]:re[7]])
	_, _ = b.Write(source[re[8]:re[9]])

	var l, u T
	var ex1, ex2 byte
	_, err := fmt.Fscanf(b, "%c%v,%v%c",
		&ex1, &l, &u, &ex2)
	if err != nil {
		return errors.WithMessagef(ParseError, "wnen scan values: %s", err)
	}

	switch ex1 {
	case '(', '[':
	default:
		panic(msgUnexpected)
	}
	switch ex2 {
	case ')', ']':
	default:
		panic(msgUnexpected)
	}
	*result = IntsFactory(T(0)).Range(l, ex1 == '(', u, ex2 == ')')
	return nil
}

var (
	intsRangeRe = regexp.MustCompile(
		`^\s*([\(\[])\s*(-?\d+)\s*,\s*(-?\d+)\s*([\)\]])\s*$`)
)
