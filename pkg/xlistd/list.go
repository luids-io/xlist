// Copyright 2020 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package xlistd

import (
	"github.com/luids-io/api/xlist"
)

// List is the main interface for RBL lists
type List interface {
	ID() string
	Class() string
	xlist.Checker
	ReadOnly() bool
}

// Finder interface for lists
type Finder interface {
	List(string) (List, bool)
}
