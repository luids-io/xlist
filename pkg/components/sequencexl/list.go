// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package sequencexl provides a simple xlist.Checker implementation that can
// be used to check in sequence on the child components.
//
// This package is a work in progress and makes no API stability promises.
package sequencexl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/luids-io/core/xlist"
)

// Config options
type Config struct {
	FirstResponse   bool
	SkipErrors      bool
	ForceValidation bool
	Reason          string
}

type options struct {
	firstResponse   bool
	skipErrors      bool
	forceValidation bool
	reason          string
}

// List implements a composite RBL that checks a group of lists in
// the order in which they were added
type List struct {
	id        string
	opts      options
	childs    []xlist.List
	provides  []bool
	resources []xlist.Resource
}

// New creates a new sequence
func New(id string, childs []xlist.List, resources []xlist.Resource, cfg Config) *List {
	l := &List{
		id: id,
		opts: options{
			firstResponse:   cfg.FirstResponse,
			skipErrors:      cfg.SkipErrors,
			forceValidation: cfg.ForceValidation,
			reason:          cfg.Reason,
		},
		resources: xlist.ClearResourceDups(resources),
		provides:  make([]bool, len(xlist.Resources), len(xlist.Resources)),
	}
	//set resource types that provides
	for _, r := range l.resources {
		l.provides[int(r)] = true
	}
	//set childs
	if len(childs) > 0 {
		l.childs = make([]xlist.List, len(childs), len(childs))
		copy(l.childs, childs)
	}
	return l
}

// ID implements xlist.List interface
func (l *List) ID() string {
	return l.id
}

// Class implements xlist.List interface
func (l *List) Class() string {
	return BuildClass
}

// Check implements xlist.Checker interface
func (l *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if !l.checks(resource) {
		return xlist.Response{}, xlist.ErrNotSupported
	}
	name, ctx, err := xlist.DoValidation(ctx, name, resource, l.opts.forceValidation)
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
		if err != nil && !l.opts.skipErrors {
			return r, err
		}
		// check if a cancellation has been done
		select {
		case <-ctx.Done():
			return xlist.Response{}, ctx.Err()
		default:
			if r.Result {
				if !result {
					result = true
					ttl = r.TTL
				} else if ttl > r.TTL {
					ttl = r.TTL
				}
				reasons = append(reasons, r.Reason)
				if l.opts.firstResponse {
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
		if l.opts.reason == "" {
			resp.Reason = strings.Join(reasons, ";")
		} else {
			resp.Reason = l.opts.reason
		}
	}
	return resp, nil
}

// Resources implements xlist.Checker interface
func (l *List) Resources() []xlist.Resource {
	resources := make([]xlist.Resource, len(l.resources), len(l.resources))
	copy(resources, l.resources)
	return resources
}

// pingResult is used for store pings
type pingResult struct {
	listIdx int
	err     error
}

// Ping implements interface xlist.Checker
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

// ReadOnly implements xlist.List interface
func (l *List) ReadOnly() bool {
	return true
}
