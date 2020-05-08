// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package grpcxl

import (
	"github.com/luids-io/core/xlist"
)

//TODO
type grpclist struct {
	id string
	xlist.Checker
}

// ID implements xlist.List interface
func (l *grpclist) ID() string {
	return l.id
}

// Class implements xlist.List interface
func (l *grpclist) Class() string {
	return BuildClass
}

// ReadOnly implements xlist.Writer interface
func (l *grpclist) ReadOnly() bool {
	return true
}
