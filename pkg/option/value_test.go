package option

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_JSON(t *testing.T) {
	type ty struct {
		Data ValueOf[string] `json:"d"`
	}
	var s, s1 ty
	data, e := json.Marshal(s)
	require.NoError(t, e)
	require.Equal(t, "{\"d\":null}", string(data))
	e = json.Unmarshal(data, &s1)
	require.NoError(t, e)
	require.Equal(t, s, s1)

	s.Data.Set("abcd")
	data, e = json.Marshal(s)
	require.NoError(t, e)
	require.Equal(t, "{\"d\":\"abcd\"}", string(data))
	e = json.Unmarshal(data, &s1)
	require.NoError(t, e)
	require.Equal(t, s, s1)
}
