// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package parallelxl provides a simple xlist.Checker implementation that can
// be used to check in parallel on the child components.
//
// This package is a work in progress and makes no API stability promises.
package parallelxl

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

// List is a composite list that does the checks in parallel
type List struct {
	xlist.List

	opts     options
	childs   []xlist.Checker
	provides []bool //slice with resources availables
}

// New returns a new parallel component with the resources passed
func New(resources []xlist.Resource, opt ...Option) *List {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	p := &List{
		opts:     opts,
		childs:   make([]xlist.Checker, 0),
		provides: make([]bool, len(xlist.Resources), len(xlist.Resources)),
	}
	for _, r := range resources {
		if r.IsValid() {
			p.provides[int(r)] = true
		}
	}
	return p
}

// AddChecker adds a checker to the RBL
func (p *List) AddChecker(list xlist.Checker) {
	p.childs = append(p.childs, list)
}

// checkResult is used for store parallel checks
type checkResult struct {
	listIdx  int
	response xlist.Response
	err      error
}

// Check implements xlist.Checker interface
func (p *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if !p.checks(resource) {
		return xlist.Response{}, xlist.ErrNotImplemented
	}
	name, ctx, err := xlist.DoValidation(ctx, name, resource, p.opts.forceValidation)
	if err != nil {
		return xlist.Response{}, err
	}
	//create context for childs
	childCtx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup
	results := make(chan *checkResult, len(p.childs))
	for idx, l := range p.childs {
		wg.Add(1)
		go workerCheck(childCtx, &wg, l, idx, name, resource, results)
	}

	ttl := 0
	result := false
	reasons := make([]string, 0, len(p.childs))
	finished := 0
RESULTLOOP:
	for finished < len(p.childs) {
		select {
		case r := <-results:
			finished++
			if r.err != nil {
				if !p.opts.skipErrors {
					err = r.err
					break RESULTLOOP
				}
			} else {
				if r.response.Result {
					if !result {
						result = true
						ttl = r.response.TTL
					} else if ttl > r.response.TTL {
						ttl = r.response.TTL
					}
					reasons = append(reasons, r.response.Reason)
					if p.opts.firstResponse {
						break RESULTLOOP
					}
				}
			}
		case <-ctx.Done():
			err = ctx.Err()
			break RESULTLOOP
		}
	}
	cancel()
	wg.Wait()
	close(results)

	var resp xlist.Response
	if result {
		resp.Result = result
		if ttl > 0 {
			resp.TTL = ttl
		}
		if p.opts.reason == "" {
			resp.Reason = strings.Join(reasons, ";")
		} else {
			resp.Reason = p.opts.reason
		}
	}
	return resp, err
}

// Resources implements xlist.Checker interface
func (p *List) Resources() []xlist.Resource {
	resources := make([]xlist.Resource, 0, len(xlist.Resources))
	for _, r := range xlist.Resources {
		if p.provides[int(r)] {
			resources = append(resources, r)
		}
	}
	return resources
}

// pingResult is used for store pings
type pingResult struct {
	listIdx int
	listKey string
	err     error
}

// Ping implements xlist.Checker interface
func (p *List) Ping() error {
	var wg sync.WaitGroup
	results := make(chan *pingResult, len(p.childs))
	for idx, l := range p.childs {
		wg.Add(1)
		go workerPing(&wg, l, idx, results)
	}

	errs := make([]*pingResult, 0, len(p.childs))
	finished := 0
	for finished < len(p.childs) {
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
			msgErr = append(msgErr, fmt.Sprintf("parallel[%v]: %v", e.listIdx, e.err))
		}
		return errors.New(strings.Join(msgErr, ";"))
	}
	return nil
}

func (p *List) checks(r xlist.Resource) bool {
	if r.IsValid() {
		return p.provides[int(r)]
	}
	return false
}

func workerCheck(ctx context.Context, wg *sync.WaitGroup, list xlist.Checker, listIdx int,
	name string, resource xlist.Resource, results chan<- *checkResult) {
	defer wg.Done()
	response, err := list.Check(ctx, name, resource)
	if err != nil {
		results <- &checkResult{
			listIdx: listIdx,
			err:     err,
		}
		return
	}
	results <- &checkResult{
		listIdx:  listIdx,
		response: response,
	}
	return
}

func workerPing(wg *sync.WaitGroup, list xlist.Checker, listIdx int, results chan<- *pingResult) {
	defer wg.Done()
	err := list.Ping()
	results <- &pingResult{
		listIdx: listIdx,
		err:     err,
	}
}

// Append implements xlist.Writer interface
func (p *List) Append(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return xlist.ErrReadOnlyMode
}

// Remove implements xlist.Writer interface
func (p *List) Remove(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return xlist.ErrReadOnlyMode
}

// Clear implements xlist.Writer interface
func (p *List) Clear(ctx context.Context) error {
	return xlist.ErrReadOnlyMode
}

// ReadOnly implements xlist.Writer interface
func (p *List) ReadOnly() (bool, error) {
	return true, nil
}
