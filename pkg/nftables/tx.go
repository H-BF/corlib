//go:build linux
// +build linux

package nftables

import (
	"net"
	"sync"

	nftlib "github.com/google/nftables"
	"github.com/pkg/errors"
	"github.com/vishvananda/netns"
)

// Tx -
type Tx struct {
	*nftlib.Conn
	commitOnce sync.Once
}

// NewTx -
func NewTx(netNS string) (*Tx, error) {
	const api = "connect to netfilter"

	opts := []nftlib.ConnOption{nftlib.AsLasting()}
	if len(netNS) > 0 {
		n, e := netns.GetFromName(netNS)
		if e != nil {
			return nil, errors.WithMessagef(e,
				"%s: accessing netns '%s'", api, netNS)
		}
		opts = append(opts, nftlib.WithNetNSFd(int(n)))
		defer n.Close()
	}
	c, e := nftlib.New(opts...)
	if e != nil {
		return nil, errors.WithMessage(e, api)
	}
	return &Tx{Conn: c}, nil
}

// Close impl 'Closer'
func (tx *Tx) Close() error {
	if tx != nil {
		c := tx.Conn
		tx.commitOnce.Do(func() {
			_ = c.CloseLasting()
		})
	}
	return nil
}

// FlushAndClose does flush and close
func (tx *Tx) FlushAndClose() error {
	c := tx.Conn
	err := net.ErrClosed
	tx.commitOnce.Do(func() {
		err = tx.Flush()
		_ = c.CloseLasting()
	})
	return err
}
