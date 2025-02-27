package dict

import (
	"encoding/json"
	"reflect"
	"sync"
	"testing"

	"github.com/H-BF/corlib/pkg/option"
	"github.com/stretchr/testify/require"
)

func Test_HDict_Json(t *testing.T) {
	var d1 HDict[string, int]
	d1.Put("A", 1)
	d1.Put("B", 2)
	d1.Put("C", 3)
	data, err := json.Marshal(d1)
	require.NoError(t, err)
	var d2 HDict[string, int]
	err = json.Unmarshal(data, &d2)
	require.NoError(t, err)
	equal := d1.Eq(&d2, func(vL, vR int) bool {
		return vL == vR
	})
	require.True(t, equal)
}

func Test_RBDict_Json(t *testing.T) {
	var d1 RBDict[string, int]
	d1.Put("A", 1)
	d1.Put("B", 2)
	d1.Put("C", 3)
	data, err := json.Marshal(d1)
	require.NoError(t, err)
	var d2 RBDict[string, int]
	err = json.Unmarshal(data, &d2)
	require.NoError(t, err)
	equal := d1.Eq(&d2, func(vL, vR int) bool {
		return vL == vR
	})
	require.True(t, equal)
}

func Test_HSet_Json(t *testing.T) {
	var s1 HSet[string]
	s1.PutMany("A", "B", "C")
	data, err := json.Marshal(s1)
	require.NoError(t, err)
	var s2 HSet[string]
	err = json.Unmarshal(data, &s2)
	require.NoError(t, err)
	equal := s1.Eq(&s2)
	require.True(t, equal)
}

func Test_RBSet_Json(t *testing.T) {
	var s1 RBSet[string]
	s1.PutMany("A", "B", "C")
	data, err := json.Marshal(s1)
	require.NoError(t, err)
	var s2 RBSet[string]
	err = json.Unmarshal(data, &s2)
	require.NoError(t, err)
	equal := s1.Eq(&s2)
	require.True(t, equal)
}

func Test_RBDict_JsonA(t *testing.T) {

	var st sync.Map
	el := reflect.ValueOf(&st).Elem()
	ty := reflect.TypeOf(sync.Map{})
	if el.CanConvert(ty) {
		i := 1
		i++
	}

	type Data struct {
		D option.ValueOf[*RBDict[string, int]]
		M string
	}
	d := Data{
		D: option.MustNewOption(&RBDict[string, int]{}),
		M: "ABC",
	}
	d.D.SomeOr(nil).Put("A", 1)
	//d.D.Set(x)

	data, err := json.Marshal(d)
	require.NoError(t, err)
	_ = data

	//var _ option.ValueOf[*RBDict[string, int]]
	d1 := Data{
		M: "ABC",
	}
	err = json.Unmarshal(data, &d1)
	require.NoError(t, err)
	i := 1
	i++

}
