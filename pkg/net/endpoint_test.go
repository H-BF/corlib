package net

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewEndpoint(t *testing.T) {
	type variant struct {
		source        string
		expectNetwork string
		expectAddress string
	}

	variants := []variant{
		{
			"  127.0.0.1:100  ",
			"tcp",
			"127.0.0.1:100",
		},
		{
			"  tcp:127.0.0.1:100  ",
			"tcp",
			"127.0.0.1:100",
		},
		{
			"  tcp://127.0.0.1:100  ",
			"tcp",
			"127.0.0.1:100",
		},
		{
			"  tcp://:100  ",
			"tcp",
			":100",
		},
		{
			"  tcp::100  ",
			"tcp",
			":100",
		},
		{
			"  :100  ",
			"tcp",
			":100",
		},
		{
			"  [::]:100  ",
			"tcp",
			"[::]:100",
		},
		{
			"  unix:~/socket  ",
			"unix",
			"~/socket",
		},
		{
			"  unix:/socket/1  ",
			"unix",
			"/socket/1",
		},
		{
			"  unix:///socket/1  ",
			"unix",
			"/socket/1",
		},
	}

	for _, v := range variants {
		ep, err := ParseEndpoint(v.source)
		assert.NoErrorf(t, err, "NewEndpoint('%s')", v.source)
		if err != nil {
			return
		}
		assert.Equalf(t, v.expectNetwork, ep.Network(), "%s", v.source)
		assert.Equalf(t, v.expectAddress, ep.String(), "%s", v.source)
	}
}
