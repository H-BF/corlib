package conventions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GrpcMethodInfo_Init(t *testing.T) {

	type caseT struct {
		source        string
		expFailed     bool
		expServiceFQN string
		expService    string
		expMethod     string
	}

	cases := []caseT{
		{
			source:        "a/b/c/crispy.healthcheck.HealthChecker/HttpCheck",
			expServiceFQN: "crispy.healthcheck.HealthChecker",
			expService:    "HealthChecker",
			expMethod:     "HttpCheck",
		},
		{
			source:        "/healthcheck.HealthChecker/HttpCheck",
			expServiceFQN: "healthcheck.HealthChecker",
			expService:    "HealthChecker",
			expMethod:     "HttpCheck",
		},
		{
			source:        "/HealthChecker/HttpCheck",
			expServiceFQN: "HealthChecker",
			expService:    "HealthChecker",
			expMethod:     "HttpCheck",
		},
		{
			source:    "/.HealthChecker/HttpCheck",
			expFailed: true,
		},
		{
			source:    "/HealthChecker./HttpCheck",
			expFailed: true,
		},
		{
			source:    "/HealthChecker/.HttpCheck",
			expFailed: true,
		},
		{
			source:    "//HttpCheck",
			expFailed: true,
		},
		{
			source:    "/.crispy.healthcheck.HealthChecker/HttpCheck",
			expFailed: true,
		},
		{
			source:    "/crispy.healthcheck.HealthChecker./HttpCheck",
			expFailed: true,
		},
	}

	for i := range cases {
		var m GrpcMethodInfo
		c := cases[i]
		err := m.Init(c.source)
		if c.expFailed {
			require.Errorf(t, err, "sample: %v", i)
			continue
		}
		require.NoErrorf(t, err, "sample: %v", i)
		require.Equalf(t, c.expServiceFQN, m.ServiceFQN, "sample: %v", i)
		require.Equalf(t, c.expService, m.Service, "sample: %v", i)
		require.Equalf(t, c.expMethod, m.Method, "sample: %v", i)
	}
}
