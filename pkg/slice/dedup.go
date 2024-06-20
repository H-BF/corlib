package slice //nolint:goimports,gofmt

import "reflect"

// DedupSlice -
func DedupSlice(sortedSlice interface{}, predicateEq func(i, j int) bool) (retLen int) {
	v := reflect.Indirect(reflect.ValueOf(sortedSlice))
	if v.Type().Kind() == reflect.Slice {
		nLen, swaper := v.Len(), reflect.Swapper(v.Interface())
		retLen = DedupAbstract(nLen, swaper, predicateEq)
		if v.CanSet() {
			v.SetLen(retLen)
		}
	}
	return
}

// DedupAbstract -
func DedupAbstract(nLen int, swapper func(i, j int), predicateEq func(i, j int) bool) (retLen int) {
	if j := 0; nLen > 0 {
		for i := 0; i < nLen; i++ {
			if !predicateEq(i, j) {
				j++
				swapper(j, i)
			}
		}
		retLen = j + 1
	}
	return
}
