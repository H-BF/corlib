package server

import (
	"encoding/json"

	"github.com/H-BF/corlib/server/internal"

	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
)

// SwaggerSpec ia alias to spec.Swagger
type SwaggerSpec = spec.Swagger

// ComposeSwaggers compose some swagger defs into one
func ComposeSwaggers(primary *SwaggerSpec, others ...*SwaggerSpec) error {
	return internal.SwaggerComposer{}.Compose(primary, others...)
}

// CloneSwaggerSpec делаем копию SwaggerSpec
func CloneSwaggerSpec(src *SwaggerSpec) (*SwaggerSpec, error) {
	const api = "clone Swagger spec"
	if src == nil {
		return nil, nil
	}
	data, err := json.Marshal(src)
	if err != nil {
		return nil, errors.Wrap(err, api)
	}
	ret := new(SwaggerSpec)
	if err = json.Unmarshal(data, ret); err != nil {
		return nil, errors.Wrap(err, api)
	}
	return ret, nil
}

var (
	_ = ComposeSwaggers
)
