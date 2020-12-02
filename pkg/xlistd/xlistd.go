// Copyright 2020 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package xlistd and subpackages implements a blacklist aggregator.
//
// This package is a work in progress and makes no API stability promises.
package xlistd

import (
	"github.com/luids-io/api/xlist"
)

// List is the main interface for RBL lists.
type List interface {
	ID() string
	Class() string
	Ping() error
	xlist.Checker
}
