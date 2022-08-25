package net

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func genTestSocketName() string {
	time.Sleep(100 * time.Millisecond)
	return path.Join("/tmp", fmt.Sprintf("test-%v-%v.socket", os.Getpid(), time.Now().Nanosecond()))
}

func Test_ListenOk(t *testing.T) {
	ep, err := ParseEndpoint("unix://" + genTestSocketName())
	if !assert.NoError(t, err) {
		return
	}
	var l net.Listener
	l, err = Listen(ep)
	if !assert.NoError(t, err) {
		return
	}
	_ = l.Close()
}

func Test_ListenFailOnFile(t *testing.T) {
	name := genTestSocketName()
	err := ioutil.WriteFile(name, []byte{0}, 0)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		_ = syscall.Unlink(name)
	}()
	var ep *Endpoint
	ep, err = ParseEndpoint("unix://" + name)
	if !assert.NoError(t, err) {
		return
	}
	_, err = Listen(ep)
	assert.Error(t, err)
}

func Test_ListenFailWhenAnotherOneListensItToo(t *testing.T) {
	name := genTestSocketName()
	ep, err := ParseEndpoint("unix://" + name)
	if !assert.NoError(t, err) {
		return
	}

	var l0 net.Listener
	l0, err = Listen(ep)
	if !assert.NoError(t, err) {
		return
	}
	defer l0.Close()
	newName := name + "-1"
	err = syscall.Rename(name, newName)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		_ = syscall.Rename(newName, name)
	}()
	var ep1 *Endpoint
	ep1, err = ParseEndpoint("unix://" + newName)
	if !assert.NoError(t, err) {
		return
	}
	var l1 net.Listener
	l1, err = Listen(ep1)
	if assert.Error(t, err) {
		return
	}
	_ = l1.Close()
}
