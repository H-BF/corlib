package plain_config

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_typeCastFunc(t *testing.T) {
	type (
		ss   string
		ii32 int32
		ii   int
		tt   time.Time
		dd   time.Duration
	)

	data := [...]func() error{
		new(typeCastFunc[UUID]).load,
		new(typeCastFunc[int]).load,
		new(typeCastFunc[uint]).load,
		new(typeCastFunc[int64]).load,
		new(typeCastFunc[uint64]).load,
		new(typeCastFunc[int32]).load,
		new(typeCastFunc[uint32]).load,
		new(typeCastFunc[int16]).load,
		new(typeCastFunc[uint16]).load,
		new(typeCastFunc[int8]).load,
		new(typeCastFunc[uint8]).load,
		new(typeCastFunc[float32]).load,
		new(typeCastFunc[float64]).load,
		new(typeCastFunc[string]).load,
		new(typeCastFunc[bool]).load,
		new(typeCastFunc[time.Time]).load,
		new(typeCastFunc[time.Duration]).load,
		new(typeCastFunc[ss]).load,
		new(typeCastFunc[tt]).load,
		new(typeCastFunc[dd]).load,
		new(typeCastFunc[ii]).load,
		new(typeCastFunc[ii32]).load,
	}
	for i := range data {
		require.NoError(t, data[i]())
	}

	var x typeCastFunc[UUID]
	x.load()
	x("cc36cb4a-08d5-45fa-9006-6421d6e35ee4")

	/*//
	assert.ErrorIs(t,
		new(typeCastFunc[struct{ a int }]).load(),
		ErrTypeCastNotSupported)
	*/
}

func Test_UUID(t *testing.T) {
	cases := []struct {
		v           any
		expect2fail bool
	}{
		{"", true},
		{uuid.UUID{}, false},
		{"cc36cb4a-08d5-45fa-9006-6421d6e35ee4", false},
		{[]byte("cc36cb4a-08d5-45fa-9006-6421d6e35ee4"), false},
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, true},
		{[15]uint8{}, true},
		{[16]uint8{}, false},
		{[17]uint8{}, true},

		{[15]int8{}, true},
		{[16]int8{}, true},
		{[17]int8{}, true},

		{0, true},
	}
	for i := range cases {
		c := cases[i]
		_, e := cast2uuid(c.v)
		if c.expect2fail {
			require.Errorf(t, e, "case #%v on value(%v)", i, c.v)
		} else {
			require.NoErrorf(t, e, "case #%v on value(%v)", i, c.v)
		}
	}
}
