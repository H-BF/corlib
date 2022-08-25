package conventions

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			source:        "/crispy.healthcheck.HealthChecker/HttpCheck",
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
		assert.Equalf(t, c.expFailed, err != nil, "sample: %v", i)
		if err != nil {
			continue
		}
		assert.Equalf(t, c.expServiceFQN, m.ServiceFQN, "sample: %v", i)
		assert.Equalf(t, c.expService, m.Service, "sample: %v", i)
		assert.Equalf(t, c.expMethod, m.Method, "sample: %v", i)
		assert.Equalf(t, c.source, m.String(), "sample: %v", i)
	}
}
