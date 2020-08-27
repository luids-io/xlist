// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package grpcxl provides a xlistd.List implementation that uses a remote
// xlist.Check service as source.
//
// This package is a work in progress and makes no API stability promises.
package grpcxl

import (
	"github.com/luids-io/api/xlist"
)

// ComponentClass registered.
const ComponentClass = "grpc"

//TODO
type grpclist struct {
	id string
	xlist.Checker
}

// ID implements xlistd.List interface.
func (l *grpclist) ID() string {
	return l.id
}

// Class implements xlistd.List interface.
func (l *grpclist) Class() string {
	return ComponentClass
}
