// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlistd

import (
	"context"
	"errors"

	"github.com/luids-io/api/xlist"
)

// ErrListRemoved is used when a list is marked as removed
var ErrListRemoved = errors.New("removed from database, update your config")

// removedList is used by builder to return lists marked as removed
type removedList struct {
	id        string
	resources []xlist.Resource
}

// ID implements xlistd.List interface.
func (d *removedList) ID() string {
	return d.id
}

// Class implements xlistd.List interface.
func (d *removedList) Class() string {
	return "deprecated"
}

// Check implements xlist.Checker.
func (d *removedList) Check(ctx context.Context, name string, res xlist.Resource) (xlist.Response, error) {
	return xlist.Response{}, nil
}

// Ping implements xlist.Checker.
func (d *removedList) Ping() error {
	return ErrListRemoved
}

// Resources implements xlist.Checker.
func (d *removedList) Resources() []xlist.Resource {
	ret := make([]xlist.Resource, len(d.resources), len(d.resources))
	copy(ret, d.resources)
	return ret
}
