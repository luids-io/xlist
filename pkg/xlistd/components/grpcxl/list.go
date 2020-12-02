// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package grpcxl provides a xlistd.List implementation that uses a remote
// xlist.Check service as source.
//
// This package is a work in progress and makes no API stability promises.
package grpcxl

import (
	"context"

	"github.com/luids-io/api/xlist"
)

// ComponentClass registered.
const ComponentClass = "grpc"

//TODO
type grpclist struct {
	id        string
	resources []xlist.Resource
	checker   xlist.Checker
}

// ID implements xlistd.List interface.
func (l *grpclist) ID() string {
	return l.id
}

// Class implements xlistd.List interface.
func (l *grpclist) Class() string {
	return ComponentClass
}

func (l *grpclist) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	return l.checker.Check(ctx, name, resource)
}

// Resources wrappes api, (it's required in construction).
func (l *grpclist) Resources(ctx context.Context) ([]xlist.Resource, error) {
	ret := make([]xlist.Resource, len(l.resources), len(l.resources))
	copy(ret, l.resources)
	return ret, nil
}

func (l *grpclist) Ping() error {
	_, err := l.checker.Resources(context.Background())
	return err
}
