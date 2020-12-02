// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package sequencexl provides a simple xlistd.List implementation that can
// be used to check in sequence on the child components.
//
// This package is a work in progress and makes no API stability promises.
package sequencexl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// ComponentClass registered.
const ComponentClass = "sequence"

// Config options
type Config struct {
	FirstResponse   bool
	SkipErrors      bool
	ForceValidation bool
	Reason          string
}

// List implements a composite RBL that checks a group of lists in
// the order in which they were added.
type List struct {
	id        string
	cfg       Config
	childs    []xlistd.List
	provides  []bool
	resources []xlist.Resource
}

// New creates a new sequence.
func New(id string, childs []xlistd.List, resources []xlist.Resource, cfg Config) *List {
	l := &List{
		id:        id,
		cfg:       cfg,
		resources: xlist.ClearResourceDups(resources, true),
		provides:  make([]bool, len(xlist.Resources), len(xlist.Resources)),
	}
	//set resource types that provides
	for _, r := range l.resources {
		l.provides[int(r)] = true
	}
	//set childs
	if len(childs) > 0 {
		l.childs = make([]xlistd.List, len(childs), len(childs))
		copy(l.childs, childs)
	}
	return l
}

// ID implements xlistd.List interface.
func (l *List) ID() string {
	return l.id
}

// Class implements xlistd.List interface.
func (l *List) Class() string {
	return ComponentClass
}

// Check implements xlist.Checker interface.
func (l *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if !l.checks(resource) {
		return xlist.Response{}, xlist.ErrNotSupported
	}
	name, ctx, err := xlist.DoValidation(ctx, name, resource, l.cfg.ForceValidation)
	if err != nil {
		return xlist.Response{}, err
	}

	// iterate over secuence list
	result := false
	ttl := 0
	reasons := make([]string, 0, len(l.childs))
LOOPCHILDS:
	for _, child := range l.childs {
		r, err := child.Check(ctx, name, resource)
		if err != nil && !l.cfg.SkipErrors {
			return r, err
		}
		// check if a cancellation has been done
		select {
		case <-ctx.Done():
			return xlist.Response{}, xlist.ErrCanceledRequest
		default:
			if r.Result {
				if !result {
					result = true
					ttl = r.TTL
				} else if ttl > r.TTL {
					ttl = r.TTL
				}
				reasons = append(reasons, r.Reason)
				if l.cfg.FirstResponse {
					break LOOPCHILDS
				}
			}
		}
	}
	var resp xlist.Response
	if result {
		resp.Result = result
		if ttl > 0 {
			resp.TTL = ttl
		}
		if l.cfg.Reason == "" {
			resp.Reason = strings.Join(reasons, ";")
		} else {
			resp.Reason = l.cfg.Reason
		}
	}
	return resp, nil
}

// Resources implements xlist.Checker interface.
func (l *List) Resources(ctx context.Context) ([]xlist.Resource, error) {
	resources := make([]xlist.Resource, len(l.resources), len(l.resources))
	copy(resources, l.resources)
	return resources, nil
}

// pingResult is used for store pings
type pingResult struct {
	listIdx int
	err     error
}

// Ping implements interface xlistd.List.
func (l *List) Ping() error {
	errs := make([]pingResult, 0, len(l.childs))
	for idx, child := range l.childs {
		err := child.Ping()
		if err != nil {
			errs = append(errs, pingResult{listIdx: idx, err: err})
		}
	}
	if len(errs) > 0 {
		msgErr := make([]string, 0, len(errs))
		for _, e := range errs {
			msgErr = append(msgErr, fmt.Sprintf("sequence[%v]: %v", e.listIdx, e.err))
		}
		return errors.New(strings.Join(msgErr, ";"))
	}
	return nil
}

func (l *List) checks(r xlist.Resource) bool {
	if r.IsValid() {
		return l.provides[int(r)]
	}
	return false
}
