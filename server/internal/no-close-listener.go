package internal

import (
	"net"
)

//NoCloseListener is net listener with fake Close meth
type NoCloseListener struct {
	net.Listener
}

//Close ...
func (nn NoCloseListener) Close() error {
	return nil
}
