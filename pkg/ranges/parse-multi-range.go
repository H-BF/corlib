package ranges

import (
	"reflect"
	"regexp"
)

type RangeParser[T any, S ParseSources] func(S) (Range[T], error)

func ParseMultiRange[T any, S ParseSources, S1 ParseSources](
	in S,
	parser RangeParser[T, S1],
	consumer func(Range[T]) bool,
) error {
	source := reflect.ValueOf(in).Convert(
		reflect.TypeOf((*[]byte)(nil)).Elem(),
	).Interface().([]byte)
	re := reMultiRange.FindAll(source, -1)
	if len(re) == 0 {
		return SourceMatchError
	}
	for _, s := range re {
		i, err := parser(S1(s))
		if err != nil {
			return err
		}
		if !consumer(i) {
			break
		}
	}
	return nil
}

var (
	reMultiRange = regexp.MustCompile(`(?:[\(\[][^\[\]\(\)]+[\)\]])`)
)
