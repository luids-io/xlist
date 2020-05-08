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

// List is a composite list that does the checks in parallel
type List struct {
	id        string
	opts      options
	childs    []xlist.List
	provides  []bool
	resources []xlist.Resource
}

// New returns a new parallel component with the resources passed
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

// AddChecker adds a checker to the RBL
func (l *List) AddChecker(list xlist.List) {
	l.childs = append(l.childs, list)
}

// checkResult is used for store parallel checks
type checkResult struct {
	listIdx  int
	response xlist.Response
	err      error
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
	//create context for childs
	childCtx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup
	results := make(chan *checkResult, len(l.childs))
	for idx, child := range l.childs {
		wg.Add(1)
		go workerCheck(childCtx, &wg, child, idx, name, resource, results)
	}

	ttl := 0
	result := false
	reasons := make([]string, 0, len(l.childs))
	finished := 0
RESULTLOOP:
	for finished < len(l.childs) {
		select {
		case r := <-results:
			finished++
			if r.err != nil {
				if !l.opts.skipErrors {
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
					if l.opts.firstResponse {
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
		if l.opts.reason == "" {
			resp.Reason = strings.Join(reasons, ";")
		} else {
			resp.Reason = l.opts.reason
		}
	}
	return resp, err
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
	listKey string
	err     error
}

// Ping implements xlist.Checker interface
func (l *List) Ping() error {
	var wg sync.WaitGroup
	results := make(chan *pingResult, len(l.childs))
	for idx, child := range l.childs {
		wg.Add(1)
		go workerPing(&wg, child, idx, results)
	}

	errs := make([]*pingResult, 0, len(l.childs))
	finished := 0
	for finished < len(l.childs) {
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
			msgErr = append(msgErr, fmt.Sprintf("%s: %v", l.childs[e.listIdx].ID(), e.err))
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

// ReadOnly implements xlist.List interface
func (l *List) ReadOnly() bool {
	return true
}
