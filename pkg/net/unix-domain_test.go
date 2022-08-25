package net

import (
	"fmt"
	"net/http"
	"path"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestFindUnixSocketFromURI(t *testing.T) {
	sockPath := genTestSocketName()
	listener, err := ListenUnixDomain(sockPath)
	if !assert.NoError(t, err) {
		return
	}
	defer listener.Close()
	uri := string(SchemeUnixHTTP) + "://" + path.Join(sockPath, "p1", "p2", "p3") + "?v1=100&v2=200"
	parsed := findUnixSocketFromURI(uri)
	assert.Equal(t, sockPath, parsed)
}

func TestUnixSocketRoundTrip(t *testing.T) {
	sockPath := genTestSocketName()
	listener, err := ListenUnixDomain(sockPath)
	if !assert.NoError(t, err) {
		return
	}
	defer listener.Close()

	m := chi.NewMux()
	uri := "/simple/get"
	m.Get(uri, func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprintf(writer, "hello there\n")
	})
	var serv http.Server
	serv.Handler = m
	go func() {
		_ = serv.Serve(listener)
	}()
	defer serv.Close()
	client := UDS.EnrichClient(http.DefaultClient)
	qry := string(SchemeUnixHTTP) + "://" + path.Join(sockPath, uri)
	var resp *http.Response
	resp, err = client.Get(qry)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotNil(t, resp) {
		return
	}
	if !assert.NotNil(t, resp) {
		return
	}
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
