package nl

import (
	"time"
)

type (
	// WithNetnsName sets net NS name
	WithNetnsName string

	WithLinger struct {
		WatcherOption
		Linger time.Duration
	}
)

type (
	// LinkListerOpt lister option
	LinkListerOpt interface {
		isLinkListerOpt()
	}

	// WatcherOption -
	WatcherOption interface {
		isWatcherOption()
	}

	scopeOfUpdates uint32
)

const (
	scopeNone scopeOfUpdates = (1 << iota) >> 1

	//IgnoreLinks does not send 'Links'
	IgnoreLinks

	//IgnoreAddress does not send 'Adresses'
	IgnoreAddress
)

func (WithNetnsName) isLinkListerOpt()  {}
func (WithNetnsName) isWatcherOption()  {}
func (WithLinger) isWatcherOption()     {}
func (scopeOfUpdates) isWatcherOption() {}
