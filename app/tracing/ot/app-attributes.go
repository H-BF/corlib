package ot

import (
	attr "go.opentelemetry.io/otel/attribute"
)

const (
	//AppBuildTsKey app build timestamp
	AppBuildTsKey = attr.Key("app.build.ts")

	//AppBuildBranchKey app source branch
	AppBuildBranchKey = attr.Key("app.build.branch")

	//AppBuildHashKey app commit hash
	AppBuildHashKey = attr.Key("app.build.hash")

	//AppBuildTagKey app source tag
	AppBuildTagKey = attr.Key("app.build.tag")
)
