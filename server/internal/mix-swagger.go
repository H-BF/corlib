package internal

import (
	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
)

// SwaggerComposer ...
type SwaggerComposer struct{}

// Compose ...
func (composer SwaggerComposer) Compose(primary *spec.Swagger, mixins ...*spec.Swagger) error {
	const api = "swagger.Compose"
	opIds := composer.getOpIds(primary)
	for _, m := range mixins {
		primary.Tags = append(primary.Tags, m.Tags...)
		for k, v := range m.Definitions {
			// assume name collisions represent IDENTICAL type. careful.
			if _, exists := primary.Definitions[k]; !exists {
				primary.Definitions[k] = v
			}
		}
		for k, v := range m.Paths.Paths {
			if _, exists := primary.Paths.Paths[k]; exists {
				return errors.Errorf("%s: the path[%s] is douplicated", api, k)
			}
			// Swagger requires that operationIds be
			// unique within a spec. If we find a
			// collision we append "Mixin0" to the
			// operatoinId we are adding, where 0 is mixin
			// index.  We assume that operationIds with
			// all the proivded specs are already unique.
			piops := composer.pathItemOps(v)
			for _, piop := range piops {
				if opIds[piop.ID] {
					return errors.Errorf("%s: operation[%s] is douplicated", api, piop.ID)
				}
				opIds[piop.ID] = true
			}
			primary.Paths.Paths[k] = v
		}
		for k, v := range m.Parameters {
			// could try to rename on conflict but would
			// have to fix $refs in the mixin. Complain
			// for now
			if _, exists := primary.Parameters[k]; !exists {
				primary.Parameters[k] = v
			}
		}
		for k, v := range m.Responses {
			// could try to rename on conflict but would
			// have to fix $refs in the mixin. Complain
			// for now
			if _, exists := primary.Responses[k]; !exists {
				primary.Responses[k] = v
			}

		}
	}
	composer.fixEmptyResponseDescriptions(primary)
	return nil
}

func (composer SwaggerComposer) fixEmptyResponseDescriptions(s *spec.Swagger) {
	for _, v := range s.Paths.Paths {
		for _, o := range [...]*spec.Operation{v.Get, v.Put, v.Post, v.Delete, v.Options, v.Head, v.Patch} {
			if o != nil {
				composer.fixEmptyDescs(o.Responses)
			}
		}
	}
	for k, v := range s.Responses {
		composer.fixEmptyDesc(&v) //nolint:gosec
		s.Responses[k] = v
	}
}

// fixEmptyDescs adds "(empty)" as the description for any Response in
// the given Responses object that doesn't already have one.
func (composer SwaggerComposer) fixEmptyDescs(rs *spec.Responses) {
	if rs == nil {
		return
	}
	composer.fixEmptyDesc(rs.Default)
	for k, v := range rs.StatusCodeResponses {
		composer.fixEmptyDesc(&v) //nolint:gosec
		rs.StatusCodeResponses[k] = v
	}
}

// fixEmptyDesc adds "(empty)" as the description to the given
// Response object if it doesn't already have one and isn't a
// ref. No-op on nil input.
func (composer SwaggerComposer) fixEmptyDesc(rs *spec.Response) {
	if rs == nil || rs.Description != "" || rs.Ref.Ref.GetURL() != nil {
		return
	}
	rs.Description = "(empty)"
}

// getOpIds extracts all the paths.<path>.operationIds from the given
// spec and returns them as the keys in a map with 'true' values.
func (composer SwaggerComposer) getOpIds(s *spec.Swagger) map[string]bool {
	rv := make(map[string]bool)
	for _, v := range s.Paths.Paths {
		for _, op := range composer.pathItemOps(v) {
			rv[op.ID] = true
		}
	}
	return rv
}

func (composer SwaggerComposer) pathItemOps(p spec.PathItem) []*spec.Operation {
	var rv []*spec.Operation
	appendOp := func(ops ...*spec.Operation) {
		for _, o := range ops {
			if o != nil {
				rv = append(rv, o)
			}
		}
	}
	appendOp(p.Get, p.Put, p.Post, p.Delete, p.Options, p.Head, p.Patch)
	return rv
}
