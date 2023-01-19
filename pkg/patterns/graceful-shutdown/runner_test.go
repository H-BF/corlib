package graceful_shutdown

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Empty(t *testing.T) {
	sta := ForDuration(time.Second).Run()
	require.Equal(t, Completed, sta)

	sta = ForDuration(0).Run()
	require.Equal(t, Timeout, sta)
}
