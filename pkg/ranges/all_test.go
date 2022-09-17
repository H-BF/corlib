package ranges

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRangesAll(t *testing.T) {
	tests := [...]struct {
		n string
		f func(*testing.T)
	}{
		{"IsTag", testBoundIsTag},
		{"CmpLowLowBounds", testCmpLowLowBounds},
		{"CmpUpUpBounds", testCmpUpUpBounds},
		{"CmpMixedBounds", testCmpMixedBounds},
		{"RangeIsNull", testRangeIsNull},
		{"BoundIsInRange", testIsIn},
		{"RangeNormalize", testNormalize},
		{"ParseIntsRange", testParseIntsRange},
		{"ParseMultiRange", testParseMultiRange},
		{"CombineMerge", testCombineMerge},
		{"CombineExclude", testCombineExclude},
		{"MultiRangeUpdate", testMultiRangeUpdate},
		{"MultiRangeSearch", testMultiRangeSearch},
	}
	for i := range tests {
		if test := tests[i]; !t.Run(test.n, test.f) {
			return
		}
		if t.Failed() {
			return
		}
	}
}

func testBoundIsTag(t *testing.T) {
	type (
		dataT = uint16
		lower = intsBound[dataT, Lower]
		upper = intsBound[dataT, Upper]
	)

	var b1 lower
	assert.True(t, b1.Is(Lower{}))
	assert.True(t, b1.Is(&Lower{}))
	assert.False(t, b1.Is(Upper{}))
	assert.False(t, b1.Is(&Upper{}))

	var b2 upper
	assert.False(t, b2.Is(Lower{}))
	assert.False(t, b2.Is(&Lower{}))
	assert.True(t, b2.Is(Upper{}))
	assert.True(t, b2.Is(&Upper{}))
}

func testCmpLowLowBounds(t *testing.T) {
	type dataT = uint16
	const dataTmax = ^dataT(0)

	factory := IntsFactory(dataTmax)
	lower := func(v dataT, ex bool) Bound[dataT] {
		return factory.Bound(Lower{}, v, ex)
	}

	testCases := [...]struct {
		A, B     Bound[dataT]
		expected int
	}{
		{lower(0, true), lower(0, false), 1},
		{lower(0, false), lower(0, true), -1},

		{lower(0, true), lower(1, false), 0},
		{lower(1, false), lower(0, true), 0},

		{lower(0, true), lower(2, false), -1},
		{lower(2, false), lower(0, true), 1},

		{lower(dataTmax, true), lower(dataTmax, false), 1},
		{lower(dataTmax, false), lower(dataTmax, true), -1},

		{lower(dataTmax-1, true), lower(dataTmax, false), 0},
		{lower(dataTmax, false), lower(dataTmax-1, true), 0},

		{lower(dataTmax-2, true), lower(dataTmax, false), -1},
		{lower(dataTmax, false), lower(dataTmax-2, true), 1},

		{lower(dataTmax, true), lower(dataTmax, true), 0},
		{lower(dataTmax, false), lower(dataTmax, false), 0},
	}
	for i, c := range testCases {
		actual := c.A.Cmp(c.B)
		require.Equalf(t, c.expected, actual,
			"%v) '%v' -cmp- '%v'", i, c.A, c.B)
	}
}

func testCmpUpUpBounds(t *testing.T) {
	type dataT = uint16
	const dataTmax = ^dataT(0)

	factory := IntsFactory(dataTmax)
	upper := func(v dataT, ex bool) Bound[dataT] {
		return factory.Bound(Upper{}, v, ex)
	}

	testCases := [...]struct {
		A, B     Bound[dataT]
		expected int
	}{
		{upper(0, true), upper(0, false), -1},
		{upper(0, false), upper(0, true), 1},

		{upper(0, false), upper(0, false), 0},
		{upper(0, true), upper(0, true), 0},

		{upper(1, true), upper(0, false), 0},
		{upper(0, false), upper(1, true), 0},

		{upper(2, true), upper(0, false), 1},
		{upper(0, false), upper(2, true), -1},

		{upper(dataTmax, false), upper(dataTmax, false), 0},
		{upper(dataTmax, true), upper(dataTmax, true), 0},

		{upper(dataTmax, false), upper(dataTmax, false), 0},
		{upper(dataTmax, true), upper(dataTmax, true), 0},

		{upper(dataTmax-1, false), upper(dataTmax, true), 0},
		{upper(dataTmax, true), upper(dataTmax-1, false), 0},

		{upper(dataTmax, false), upper(dataTmax, true), 1},
		{upper(dataTmax, true), upper(dataTmax, false), -1},

		{upper(dataTmax-2, false), upper(dataTmax, true), -1},
		{upper(dataTmax, true), upper(dataTmax-2, false), 1},
	}
	for i, c := range testCases {
		actual := c.A.Cmp(c.B)
		require.Equalf(t, c.expected, actual,
			"%v) '%v' -cmp- '%v'", i, c.A, c.B)
	}
}

func testCmpMixedBounds(t *testing.T) {
	type dataT = uint16

	factory := IntsFactory[dataT](0)
	lower := func(v dataT, ex bool) Bound[dataT] {
		return factory.Bound(Lower{}, v, ex)
	}
	upper := func(v dataT, ex bool) Bound[dataT] {
		return factory.Bound(Upper{}, v, ex)
	}

	testCases := [...]struct {
		A        Bound[dataT]
		B        Bound[dataT]
		expected int
	}{
		{lower(0, true), upper(0, true), 1},
		{upper(0, true), lower(0, true), -1},

		{lower(0, false), upper(0, true), 1},
		{upper(0, true), lower(0, false), -1},

		{lower(0, true), upper(0, false), 1},
		{upper(0, false), lower(0, true), -1},

		{lower(0, false), upper(0, false), 0},
		{upper(0, false), lower(0, false), 0},
		{lower(0, false), upper(1, true), 0},
		{lower(0, true), upper(1, false), 0},
		{upper(1, true), lower(0, false), 0},
		{upper(1, false), lower(0, true), 0},

		{lower(0, true), upper(2, true), 0},
		{upper(2, true), lower(0, true), 0},

		{lower(1, false), upper(2, true), 0},
		{lower(1, false), upper(2, false), -1},

		{upper(1, true), lower(2, true), -1},
		{upper(2, true), lower(2, true), -1},
		{upper(3, true), lower(2, true), -1},
		{upper(4, true), lower(2, true), 0},
		{upper(5, true), lower(2, true), 1},

		{upper(1, false), lower(2, true), -1},
		{upper(2, false), lower(2, true), -1},
		{upper(3, false), lower(2, true), 0},
		{upper(4, false), lower(2, true), 1},
		{upper(5, false), lower(2, true), 1},

		{upper(1, true), upper(1, true).Adjacent(), -1},
		{lower(1, false), upper(1, true).Adjacent(), 0},
		{lower(1, false).Adjacent(), upper(1, true), 0},
		{lower(1, true).Adjacent(), upper(1, false), 0},
	}
	for i, c := range testCases {
		actual := c.A.Cmp(c.B)
		require.Equalf(t, c.expected, actual,
			"%v) '%s' -cmp- '%s'", i, c.A, c.B)
	}
}

func testRangeIsNull(t *testing.T) {
	type (
		dataT = uint16
		IVal  = Range[dataT]
	)
	const dataTmax = ^dataT(0)

	factory := IntsFactory[dataT](0)
	newRange := factory.Range
	cases := [...]struct {
		val      IVal
		expected bool
	}{
		{newRange(0, false, 0, false), false},
		{newRange(0, true, 0, true), true},
		{newRange(0, true, 0, false), true},
		{newRange(0, false, 0, true), true},

		{newRange(dataTmax, false, dataTmax, false), false},
		{newRange(dataTmax, true, dataTmax, true), true},
		{newRange(dataTmax, false, dataTmax, true), true},
		{newRange(dataTmax, true, dataTmax, false), true},

		{newRange(dataTmax, false, dataTmax-1, false), true},
		{newRange(dataTmax, true, dataTmax-1, true), true},
		{newRange(dataTmax, false, dataTmax-1, true), true},
		{newRange(dataTmax, true, dataTmax-1, false), true},

		{newRange(dataTmax-1, false, dataTmax, false), false},
		{newRange(dataTmax-1, false, dataTmax, true), false},
		{newRange(dataTmax-1, true, dataTmax, false), false},
		{newRange(dataTmax-1, true, dataTmax, true), true},

		{newRange(13, false, 13, true), true},
		{newRange(13, true, 13, false), true},
		{newRange(13, true, 13, true), true},
		{newRange(13, false, 13, false), false},
	}
	for i, c := range cases {
		actual := c.val.IsNull()
		require.Equalf(t, c.expected, actual,
			"%v) is-null %s", i, c.val)
	}
}

func testIsIn(t *testing.T) {
	type dataT = uint16

	factory := IntsFactory[dataT](0)
	lower := func(v dataT, ex bool) Bound[dataT] {
		return factory.Bound(Lower{}, v, ex)
	}
	upper := func(v dataT, ex bool) Bound[dataT] {
		return factory.Bound(Upper{}, v, ex)
	}
	newRange := factory.Range
	casesL := [...]struct {
		val      Range[dataT]
		b        Bound[dataT]
		expected bool
	}{
		{newRange(10, false, 20, false), lower(9, false), false},
		{newRange(10, false, 20, false), lower(9, true), true},
		{newRange(10, true, 20, false), lower(10, false), false},
		{newRange(10, true, 20, false), lower(10, true), true},
		{newRange(10, true, 20, false), lower(11, false), true},

		{newRange(10, true, 20, false), lower(20, false), true},
		{newRange(10, true, 20, false), lower(21, false), false},
		{newRange(10, true, 20, false), lower(19, true), true},

		{newRange(10, false, 20, true), lower(19, false), true},
		{newRange(10, false, 20, true), lower(19, true), false},
		{newRange(10, false, 20, true), lower(18, true), true},
	}
	for i, c := range casesL {
		actual := c.b.IsIn(c.val)
		require.Equalf(t, c.expected, actual,
			"%v) '%s' is-in '%s'", i, c.b, c.val)
	}

	casesU := [...]struct {
		val      Range[dataT]
		b        Bound[dataT]
		expected bool
	}{
		{newRange(10, false, 20, false), upper(9, false), false},
		{newRange(10, false, 20, false), upper(10, false), true},
		{newRange(10, false, 20, false), upper(10, true), false},
		{newRange(10, false, 20, false), upper(11, true), true},

		{newRange(10, true, 20, false), upper(10, false), false},
		{newRange(10, true, 20, false), upper(11, false), true},
		{newRange(10, true, 20, false), upper(11, true), false},
		{newRange(10, true, 20, false), upper(12, true), true},

		{newRange(10, false, 20, false), upper(20, false), true},
		{newRange(10, false, 20, true), upper(20, true), true},
		{newRange(10, false, 20, false), upper(20, true), true},
		{newRange(10, false, 20, true), upper(19, false), true},
		{newRange(10, false, 20, false), upper(21, true), true},
	}
	for i, c := range casesU {
		actual := c.b.IsIn(c.val)
		require.Equalf(t, c.expected, actual,
			"%v) '%s' is-in '%s'", i, c.b, c.val)
	}
}

func testNormalize(t *testing.T) {
	type dataT = uint16

	const dataTmax = ^dataT(0)

	factory := IntsFactory[dataT](0)
	NR := factory.Range

	cases := []struct {
		v        Range[dataT]
		expected string
	}{
		//(10, 15] -> [11, 16)
		{NR(10, true, 15, false), "[11,16)"},

		//(10,10] -> (10,10]
		{NR(10, true, 10, false), "(10,10]"},

		//[10, 10) -> [10, 10)
		{NR(10, false, 10, true), "[10,10)"},

		//[10, 10] -> [10, 11)
		{NR(10, false, 10, false), "[10,11)"},

		//(10, 10) -> (10, 10)
		{NR(10, true, 10, true), "(10,10)"},

		{NR(dataTmax, false, dataTmax, false),
			NR(dataTmax, false, dataTmax, false).String()},
	}
	for i, c := range cases {
		v := c.v.Normalize()
		require.Equalf(t, c.expected, v.String(),
			"%v) Normalize('%s')", i, c.v)
	}
}

func testCombineMerge(t *testing.T) {
	type (
		dataT = uint8
	)
	cases := []struct {
		c        string
		expected string
	}{
		{"[8, 7][10, 20]", "[10,21)"},
		{"[10, 20], [8,7]", "[10,21)"},

		{"[10,20][12,21][19,22]", "[10,23)"},
		{"[19,22][10,21)[12,22)", "[10,23)"},

		{"[1,10][2,15][17,17)", "[1,16)"},
		{"[1,10] [6,15] [7,9)  [17, 17) [17, 22]", "[1,16)[17,23)"},
	}
	buf := bytes.NewBuffer(nil)
	var values []Range[dataT]
	for i := range cases { //nolint
		values = values[:0]
		c := cases[i]
		err := ParseMultiRange(c.c,
			func(src []byte) (Range[dataT], error) {
				var ret Range[dataT]
				e := ParseIntsRange(src, &ret)
				return ret, e
			},
			func(i Range[dataT]) bool {
				values = append(values, i)
				return true
			})
		require.NoErrorf(t, err, "%v) on parsing '%s'", i, cases[i].c)
		factory := IntsFactory(dataT(0))
		CombineRanges(
			func(l, u Bound[dataT]) Range[dataT] {
				ret := factory.Range(0, false, 0, true)
				ret.SetBounds(l, u)
				return ret
			},
			func(v Range[dataT]) bool {
				_, err = fmt.Fprintf(buf, "%s", v)
				require.NoError(t, err)
				return true
			}, CombineMerge, values...)
		require.Equalf(t, c.expected, buf.String(), "%v) case '%s'", i, c.c)
		buf.Reset()
	}
}

func testCombineExclude(t *testing.T) { //nolint
	type dataT = uint8

	cases := []struct {
		c        string
		expected string
	}{

		{"[10,9]", ""},
		{"(10,10)", ""},
		{"(11,11)(10,10)", ""},
		{"(11,11)(22,30)", "[23,30)"},
		{"(22,30)(33,33)", "[23,30)"},

		{"[10,20)[13,16)", "[10,13)[16,20)"},
		{"[10,20)[13,16]", "[10,13)[17,20)"},
		{"[10,20)[13,22)", "[10,13)[20,22)"},

		{"[10,20][10,21]", "[21,22)"},
		{"[10,15][10,16][10,18]", "[17,19)"},
		{"[10,20][12,22][13,30]", "[10,12)[23,31)"},

		{"[10,20][12,15][13,30][14,14)", "[10,12)[21,31)"},
		{"[10,20][12,15][13,30]", "[10,12)[21,31)"},

		{"[10,20][12,15][13,30][35,40]", "[10,12)[21,31)[35,41)"},

		{"[10,30][10,15]", "[16,31)"},
		{"[10,30][10,15][10,20][10,22]", "[23,31)"},

		////
		{"[1,3][5,8]", "[1,4)[5,9)"},
		{"[18,30][10,30]", "[10,18)"},
		{"[18,30][10,30][22,30][25,30]", "[10,18)"},
		{"[1,30][4,8][11,15]", "[1,4)[9,11)[16,31)"},
		{"[1,30][10,18][12,20]", "[1,10)[21,31)"},
		{"[2,20][10,22][12,30]", "[2,10)[23,31)"},
		{"[2,20][10,22][12,30][31,33]", "[2,10)[23,31)[31,34)"},
		{"[1,2)[2,20][10,22][12,30][31,33]", "[1,2)[2,10)[23,31)[31,34)"},

		{"[18,30][10,30][10,30][18,30]", ""},

		{"[18,30][10,30][40,50][40,55]", "[10,18)[51,56)"},

		{"[1,40][5,10],[12,18]", "[1,5)[11,12)[19,41)"},
		{"[1,40][5,10],[12,18][13,22]", "[1,5)[11,12)[23,41)"},

		{"[1,20][20,30]", "[1,20)[21,31)"},
	}
	buf := bytes.NewBuffer(nil)
	var values []Range[dataT]
	for i := range cases { //nolint
		values = values[:0]
		c := cases[i]
		err := ParseMultiRange(c.c,
			func(src []byte) (Range[dataT], error) {
				var ret Range[dataT]
				e := ParseIntsRange(src, &ret)
				return ret, e
			},
			func(i Range[dataT]) bool {
				values = append(values, i)
				return true
			})
		require.NoErrorf(t, err, "%v) on parsing '%s'", i, cases[i].c)
		factory := IntsFactory(dataT(0))
		CombineRanges(
			func(l, u Bound[dataT]) Range[dataT] {
				ret := factory.Range(0, false, 0, true)
				ret.SetBounds(l, u)
				return ret
			},
			func(v Range[dataT]) bool {
				_, err = fmt.Fprintf(buf, "%s", v)
				require.NoError(t, err)
				return true
			}, CombineExclude, values...)
		require.Equalf(t, c.expected, buf.String(), "%v) case '%s'", i, c.c)
		buf.Reset()
	}
}

func testParseIntsRange(t *testing.T) {
	type dataT = uint8

	cases := []struct {
		src      string
		isError  bool
		expected string
	}{
		{" ( 10 , 10 ) ", false, "(10,10)"},
		{" ( 10 , 10 ] ", false, "(10,10]"},
		{" [ 10 , 10 ] ", false, "[10,10]"},
		{" [ 10 , 10 ) ", false, "[10,10)"},

		{"[ [ 10 , 10 ) ", true, ""},
		{" [ 10 , 10 ( ", true, ""},
		{" [ 10 , a10 ] ", true, ""},
		{" [ -10 , 10 ] ", true, ""},
		{" [ 10 , 1000 ] ", true, ""},
		{" [ -10  10 ] ", true, ""},
	}

	for i, c := range cases {
		var v Range[dataT]
		err := ParseIntsRange(c.src, &v)
		if c.isError {
			require.Errorf(t, err, "%v) SRC='%s'", i, c.src)
		} else {
			require.Equalf(t, c.expected, v.String(), "%v) SRC='%s'", i, c.src)
		}
	}
}

func testParseMultiRange(t *testing.T) {
	type dataT = uint16

	const src = `
  {[10 ,  11] , (10,15)   ,  jh [1,12) (() [] ,m
(0,5][1 ,10  )}
`
	buf := bytes.NewBuffer(nil)
	buf2 := bytes.NewBuffer(nil)
	err := ParseMultiRange(src,
		func(src []byte) (Range[dataT], error) {
			var r Range[dataT]
			e := ParseIntsRange(src, &r)
			return r, e
		},
		func(i Range[dataT]) bool {
			if _, e := fmt.Fprintf(buf, "%s", i); assert.NoError(t, e) {
				_, e = fmt.Fprintf(buf2, "%s", i.Normalize())
				return assert.NoError(t, e)
			}
			return false
		})
	require.NoError(t, err)
	require.Equal(t, "[10,11](10,15)[1,12)(0,5][1,10)", buf.String())
	require.Equal(t, "[10,12)[11,15)[1,12)[1,6)[1,10)", buf2.String())
}

func testMultiRangeUpdate(t *testing.T) {
	tests := [...]struct {
		src string
		s   CombineStrategy
		exp string
	}{
		{"[10,20][12,15][13,30][35,40]", CombineExclude, "[10,12)[21,31)[35,41)"},
		{"[1,10] [6,15] [7,9)  [17, 17) [17, 22]", CombineMerge, "[1,16)[17,23)"},
	}
	type typeT = uint32
	f := IntsFactory[typeT](0)
	for i := range tests {
		c := tests[i]

		var rgs []Range[typeT]
		err := ParseMultiRange(c.src,
			func(src []byte) (Range[typeT], error) {
				var r Range[typeT]
				e := ParseIntsRange(src, &r)
				return r, e
			},
			func(i Range[typeT]) bool {
				rgs = append(rgs, i)
				return true
			})
		require.NoError(t, err)
		mr := NewMultiRange(f)
		mr.Update(c.s, rgs...)
		require.Equal(t, c.exp, mr.String())
	}
}

func testMultiRangeSearch(t *testing.T) {
	type typeT = uint32

	const (
		rangeCount          = 10000
		rangeInterval typeT = 10
		rangeLen      typeT = 20
	)

	f := IntsFactory[typeT](0)
	rr := make([]Range[typeT], 0, rangeCount)
	var k typeT
	for i := 0; i < rangeCount; i++ {
		rr = append(rr, f.Range(k, false, k+rangeLen, false))
		k += rangeLen + rangeInterval
	}
	rand.Shuffle(len(rr), func(i, j int) {
		rr[i], rr[j] = rr[j], rr[i]
	})
	mr := NewMultiRange(f)
	mr.Update(CombineMerge, rr...)
	require.Equal(t, len(rr), mr.Len())

	cases := [...]struct {
		pt            typeT
		shouldBeFound bool
	}{
		{0, true},
		{5, true},
		{rangeLen, true},
		{rangeLen + 1, false},
		{rangeLen + 2, false},
		{rangeLen + 3, false},
	}

	for _, c := range cases {
		pt := c.pt
		for i := 0; i < rangeCount; i++ {
			n, ok := mr.Search(pt)
			require.Equal(t, c.shouldBeFound, ok)
			if c.shouldBeFound {
				require.True(t, mr.At(n).Contains(pt))
			}
			pt += rangeLen + rangeInterval
		}
	}
}
