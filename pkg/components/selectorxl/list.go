// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package selectorxl provides a simple xlist.Checker implementation that can
// be used to select the component defined to the requested resource.
//
// This package is a work in progress and makes no API stability promises.
package selectorxl

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/luids-io/core/xlist"
)

type options struct {
	firstResponse   bool
	forceValidation bool
	reason          string
}

var defaultOptions = options{}

// Option is used for common component configuration
type Option func(*options)

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

// List is a composite list that redirects requests to RBLs based on the
// resource type.
type List struct {
	xlist.List
	opts     options
	checkers map[xlist.Resource]xlist.List
}

// New returns a new selector component
func New(opt ...Option) *List {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	s := &List{
		opts:     opts,
		checkers: make(map[xlist.Resource]xlist.List),
	}
	return s
}

//SetService sets a blacklist interface for a resource type and enables it
func (s *List) SetService(resource xlist.Resource, list xlist.List) *List {
	if !resource.IsValid() {
		return s //do noting
	}
	s.checkers[resource] = list
	return s
}

// Check implements xlist.Checker interface
func (s *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	list, ok := s.checkers[resource]
	if !ok {
		return xlist.Response{}, xlist.ErrNotImplemented
	}
	name, ctx, err := xlist.DoValidation(ctx, name, resource, s.opts.forceValidation)
	if err != nil {
		return xlist.Response{}, err
	}
	resp, err := list.Check(ctx, name, resource)
	if err == nil && resp.Result && s.opts.reason != "" {
		resp.Reason = s.opts.reason
	}
	return resp, err
}

// Resources implements xlist.Checker interface
func (s *List) Resources() []xlist.Resource {
	ret := make([]xlist.Resource, 0, len(xlist.Resources))
	for _, r := range xlist.Resources {
		_, ok := s.checkers[r]
		if ok {
			ret = append(ret, r)
		}
	}
	return ret
}

// pingResult is used for store pings
type pingResult struct {
	listIdx int
	listKey string
	err     error
}

// Ping implements xlist.Checker interface
func (s *List) Ping() error {
	var wg sync.WaitGroup
	results := make(chan *pingResult, len(s.checkers))
	for key, l := range s.checkers {
		wg.Add(1)
		go workerPing(&wg, l, key.String(), results)
	}
	errs := make([]*pingResult, 0, len(s.checkers))
	finished := 0
	for finished < len(s.checkers) {
		select {
		case result := <-results:
			finished++
			if result.err != nil {
				errs = append(errs, result)
			}
		}
	}
	wg.Wait()
	close(results)

	if len(errs) > 0 {
		msgErr := make([]string, 0, len(errs))
		for _, e := range errs {
			msgErr = append(msgErr, fmt.Sprintf("selector[%v]: %v", e.listKey, e.err))
		}
		return errors.New(strings.Join(msgErr, ";"))
	}
	return nil
}

func workerPing(wg *sync.WaitGroup, list xlist.Checker, listKey string,
	results chan<- *pingResult) {
	defer wg.Done()
	err := list.Ping()
	results <- &pingResult{
		listKey: listKey,
		err:     err,
	}
}

// Append implements xlist.Writer interface
func (s *List) Append(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return xlist.ErrReadOnlyMode
}

// Remove implements xlist.Writer interface
func (s *List) Remove(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return xlist.ErrReadOnlyMode
}

// Clear implements xlist.Writer interface
func (s *List) Clear(ctx context.Context) error {
	return xlist.ErrReadOnlyMode
}

// ReadOnly implements xlist.Writer interface
func (s *List) ReadOnly() (bool, error) {
	return true, nil
}
