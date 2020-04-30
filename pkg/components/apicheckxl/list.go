// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package apicheckxl

import (
	"context"

	"github.com/luids-io/core/xlist"
)

type apicheckList struct {
	resources []xlist.Resource
	xlist.Checker
}

// Resources wrappes api, is required in construction
func (l *apicheckList) Resources() []xlist.Resource {
	ret := make([]xlist.Resource, len(l.resources), len(l.resources))
	copy(ret, l.resources)
	return ret
}

// Append implements xlist.Writer interface
func (l *apicheckList) Append(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return xlist.ErrReadOnlyMode
}

// Remove implements xlist.Writer interface
func (l *apicheckList) Remove(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return xlist.ErrReadOnlyMode
}

// Clear implements xlist.Writer interface
func (l *apicheckList) Clear(ctx context.Context) error {
	return xlist.ErrReadOnlyMode
}

// ReadOnly implements xlist.Writer interface
func (l *apicheckList) ReadOnly() (bool, error) {
	return true, nil
}
