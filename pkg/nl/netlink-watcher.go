//go:build !linux
// +build !linux

package nl

import (
	"github.com/pkg/errors"
)

// NewNetlinkWatcher creates NetlinkWatcher instance
func NewNetlinkWatcher(opts ...WatcherOption) (NetlinkWatcher, error) {
	return nil, errors.New("NewNetlinkWatcher: NotImpl")
}
