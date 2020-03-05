// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package grpcxl

import (
	"context"

	"github.com/luids-io/core/xlist"
)

//TODO
type grpclist struct {
	xlist.Checker
}

// Append implements xlist.Writer interface
func (l *grpclist) Append(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return xlist.ErrReadOnlyMode
}

// Remove implements xlist.Writer interface
func (l *grpclist) Remove(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return xlist.ErrReadOnlyMode
}

// Clear implements xlist.Writer interface
func (l *grpclist) Clear(ctx context.Context) error {
	return xlist.ErrReadOnlyMode
}

// ReadOnly implements xlist.Writer interface
func (l *grpclist) ReadOnly() (bool, error) {
	return true, nil
}
