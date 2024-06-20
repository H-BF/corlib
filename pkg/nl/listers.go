//go:build !linux
// +build !linux

package nl

import (
	"context"

	"github.com/pkg/errors"
)

// NewLinkLister -
func NewLinkLister(ctx context.Context, opts ...LinkListerOpt) (LinkLister, error) {
	return nil, errors.New("NewLinkLister: NotImpl")
}
