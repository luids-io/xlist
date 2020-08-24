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

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// ComponentClass registered
const ComponentClass = "selector"

// Config options
type Config struct {
	ForceValidation bool
	Reason          string
}

type options struct {
	forceValidation bool
	reason          string
}

// List is a composite list that redirects requests to RBLs based on the
// resource type.
type List struct {
	id        string
	opts      options
	services  []xlistd.List
	resources []xlist.Resource
}

// New returns a new selector component
func New(id string, services map[xlist.Resource]xlistd.List, cfg Config) *List {
	l := &List{
		id: id,
		opts: options{
			forceValidation: cfg.ForceValidation,
			reason:          cfg.Reason,
		},
		services:  make([]xlistd.List, len(xlist.Resources), len(xlist.Resources)),
		resources: make([]xlist.Resource, 0, len(services)),
	}
	for res, list := range services {
		if res.IsValid() {
			if l.services[int(res)] == nil {
				l.services[int(res)] = list
				l.resources = append(l.resources, res)
			}
		}
	}
	l.resources = xlist.ClearResourceDups(l.resources)
	return l
}

// ID implements xlistd.List interface
func (l *List) ID() string {
	return l.id
}

// Class implements xlistd.List interface
func (l *List) Class() string {
	return ComponentClass
}

// Check implements xlist.Checker interface
func (l *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	list := l.getList(resource)
	if list == nil {
		return xlist.Response{}, xlist.ErrNotSupported
	}
	name, ctx, err := xlist.DoValidation(ctx, name, resource, l.opts.forceValidation)
	if err != nil {
		return xlist.Response{}, err
	}
	resp, err := list.Check(ctx, name, resource)
	if err == nil && resp.Result && l.opts.reason != "" {
		resp.Reason = l.opts.reason
	}
	return resp, err
}

// Resources implements xlist.Checker interface
func (l *List) Resources() []xlist.Resource {
	ret := make([]xlist.Resource, len(l.resources), len(l.resources))
	copy(ret, l.resources)
	return ret
}

func (l *List) getList(r xlist.Resource) xlistd.List {
	if r.IsValid() {
		return l.services[int(r)]
	}
	return nil
}

// pingResult is used for store pings
type pingResult struct {
	res xlist.Resource
	err error
}

// Ping implements xlist.Checker interface
func (l *List) Ping() error {
	var wg sync.WaitGroup
	results := make(chan *pingResult, len(l.resources))
	for _, res := range l.resources {
		child := l.services[int(res)]
		if child != nil {
			wg.Add(1)
			go workerPing(&wg, child, res, results)
		}
	}
	errs := make([]*pingResult, 0, len(l.resources))
	finished := 0
	for finished < len(l.resources) {
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
			msgErr = append(msgErr, fmt.Sprintf("%s: %v", l.services[e.res].ID(), e.err))
		}
		return errors.New(strings.Join(msgErr, ";"))
	}
	return nil
}

func workerPing(wg *sync.WaitGroup, list xlist.Checker, res xlist.Resource, results chan<- *pingResult) {
	defer wg.Done()
	err := list.Ping()
	results <- &pingResult{
		res: res,
		err: err,
	}
}

// ReadOnly implements xlistd.List interface
func (l *List) ReadOnly() bool {
	return true
}
