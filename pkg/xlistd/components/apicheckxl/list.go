// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package apicheckxl provides a xlistd.List implementation that uses an
// xlist.Check api service for its checks.
//
// This package is a work in progress and makes no API stability promises.
package apicheckxl

import (
	"github.com/luids-io/api/xlist"
)

// ComponentClass registered.
const ComponentClass = "apicheck"

type apicheckList struct {
	id        string
	resources []xlist.Resource
	xlist.Checker
}

func (l *apicheckList) ID() string {
	return l.id
}

func (l *apicheckList) Class() string {
	return ComponentClass
}

// Resources wrappes api, (it's required in construction).
func (l *apicheckList) Resources() []xlist.Resource {
	ret := make([]xlist.Resource, len(l.resources), len(l.resources))
	copy(ret, l.resources)
	return ret
}
