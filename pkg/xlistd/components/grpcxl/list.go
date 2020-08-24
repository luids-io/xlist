// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package grpcxl

import (
	"github.com/luids-io/api/xlist"
)

// ComponentClass registered
const ComponentClass = "grpc"

//TODO
type grpclist struct {
	id string
	xlist.Checker
}

// ID implements xlistd.List interface
func (l *grpclist) ID() string {
	return l.id
}

// Class implements xlistd.List interface
func (l *grpclist) Class() string {
	return ComponentClass
}

// ReadOnly implements xlist.Writer interface
func (l *grpclist) ReadOnly() bool {
	return true
}
