// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package apicheckxl

import (
	"github.com/luids-io/core/xlist"
)

type apicheckList struct {
	id        string
	resources []xlist.Resource
	xlist.Checker
}

func (l *apicheckList) ID() string {
	return l.id
}

func (l *apicheckList) Class() string {
	return BuildClass
}

// Resources wrappes api, is required in construction
func (l *apicheckList) Resources() []xlist.Resource {
	ret := make([]xlist.Resource, len(l.resources), len(l.resources))
	copy(ret, l.resources)
	return ret
}

// ReadOnly implements xlist.Writer interface
func (l *apicheckList) ReadOnly() bool {
	return true
}