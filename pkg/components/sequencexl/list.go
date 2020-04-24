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
	Resources       []xlist.Resource
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
	opts      options
	childs    []xlist.Checker
	provides  []bool
	resources []xlist.Resource
}

// New creates a new sequence
func New(childs []xlist.Checker, cfg Config) *List {
	l := &List{
		opts: options{
			firstResponse:   cfg.FirstResponse,
			skipErrors:      cfg.SkipErrors,
			forceValidation: cfg.ForceValidation,
			reason:          cfg.Reason,
		},
		resources: xlist.ClearResourceDups(cfg.Resources),
		provides:  make([]bool, len(xlist.Resources), len(xlist.Resources)),
	}
	//set resource types that provides
	for _, r := range l.resources {
		l.provides[int(r)] = true
	}
	//set childs
	if len(childs) > 0 {
		l.childs = make([]xlist.Checker, len(childs), len(childs))
		copy(l.childs, childs)
	}
	return l
}

// Check implements xlist.Checker interface
func (l *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if !l.checks(resource) {
		return xlist.Response{}, xlist.ErrNotImplemented
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

// Append implements xlist.Writer interface
func (l *List) Append(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return xlist.ErrReadOnlyMode
}

// Remove implements xlist.Writer interface
func (l *List) Remove(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return xlist.ErrReadOnlyMode
}

// Clear implements xlist.Writer interface
func (l *List) Clear(ctx context.Context) error {
	return xlist.ErrReadOnlyMode
}

// ReadOnly implements xlist.Writer interface
func (l *List) ReadOnly() (bool, error) {
	return true, nil
}
