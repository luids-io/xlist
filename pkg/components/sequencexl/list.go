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

type options struct {
	firstResponse   bool
	skipErrors      bool
	forceValidation bool
	reason          string
}

var defaultOptions = options{}

// Option is used for common component configuration
type Option func(*options)

// FirstResponse return first child with positive result
func FirstResponse(b bool) Option {
	return func(o *options) {
		o.firstResponse = b
	}
}

// ForceValidation forces components to ignore context and validate requests
func ForceValidation(b bool) Option {
	return func(o *options) {
		o.forceValidation = b
	}
}

// Reason sets a fixed reason for component
func Reason(s string) Option {
	return func(o *options) {
		o.reason = s
	}
}

// SkipErrors option skips errors produced by childs
func SkipErrors(b bool) Option {
	return func(o *options) {
		o.skipErrors = b
	}
}

// List implements a composite RBL that checks a group of lists in
// the order in which they were added
type List struct {
	opts     options
	childs   []xlist.Checker
	provides []bool //slice with resources available
}

// New creates a new sequence
func New(resources []xlist.Resource, opt ...Option) *List {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	s := &List{
		opts:     opts,
		childs:   make([]xlist.Checker, 0),
		provides: make([]bool, len(xlist.Resources), len(xlist.Resources)),
	}
	for _, r := range resources {
		if r.IsValid() {
			s.provides[int(r)] = true
		}
	}
	return s
}

// Append adds a RBL to the sequence, if stopOnError then the checks and pings will
// return an error if something goes wrong.
func (s *List) Append(list xlist.Checker) {
	s.childs = append(s.childs, list)
}

// Check implements xlist.Checker interface
func (s *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if !s.checks(resource) {
		return xlist.Response{}, xlist.ErrResourceNotSupported
	}
	name, ctx, err := xlist.DoValidation(ctx, name, resource, s.opts.forceValidation)
	if err != nil {
		return xlist.Response{}, err
	}

	// iterate over secuence list
	result := false
	ttl := 0
	reasons := make([]string, 0, len(s.childs))
LOOPCHILDS:
	for _, l := range s.childs {
		r, err := l.Check(ctx, name, resource)
		if err != nil && !s.opts.skipErrors {
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
				if s.opts.firstResponse {
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
		if s.opts.reason == "" {
			resp.Reason = strings.Join(reasons, ";")
		} else {
			resp.Reason = s.opts.reason
		}
	}
	return resp, nil
}

// Resources implements xlist.Checker interface
func (s *List) Resources() []xlist.Resource {
	resources := make([]xlist.Resource, 0, len(xlist.Resources))
	for _, r := range xlist.Resources {
		if s.provides[int(r)] {
			resources = append(resources, r)
		}
	}
	return resources
}

// pingResult is used for store pings
type pingResult struct {
	listIdx int
	err     error
}

// Ping implements interface xlist.Checker
func (s *List) Ping() error {
	errs := make([]pingResult, 0, len(s.childs))
	for idx, l := range s.childs {
		err := l.Ping()
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

func (s *List) checks(r xlist.Resource) bool {
	if r.IsValid() {
		return s.provides[int(r)]
	}
	return false
}
