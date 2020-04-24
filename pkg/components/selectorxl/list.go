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
	opts      options
	services  []xlist.List
	resources []xlist.Resource
	readonly  bool
}

// New returns a new selector component
func New(services map[xlist.Resource]xlist.List, cfg Config) *List {
	l := &List{
		opts: options{
			forceValidation: cfg.ForceValidation,
			reason:          cfg.Reason,
		},
		services:  make([]xlist.List, len(xlist.Resources), len(xlist.Resources)),
		resources: make([]xlist.Resource, 0, len(services)),
	}
	for res, list := range services {
		if readonly, _ := list.ReadOnly(); readonly {
			l.readonly = true
		}
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

// Check implements xlist.Checker interface
func (l *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	list := l.getList(resource)
	if list == nil {
		return xlist.Response{}, xlist.ErrNotImplemented
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

func (l *List) getList(r xlist.Resource) xlist.List {
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
			msgErr = append(msgErr, fmt.Sprintf("selector[%v]: %v", e.res, e.err))
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

// Append implements xlist.Writer interface
func (l *List) Append(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	if !l.readonly {
		if list := l.getList(r); list != nil {
			return list.Append(ctx, name, r, f)
		}
		return xlist.ErrNotAvailable
	}
	return xlist.ErrReadOnlyMode
}

// Remove implements xlist.Writer interface
func (l *List) Remove(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	if !l.readonly {
		if list := l.getList(r); list != nil {
			return list.Remove(ctx, name, r, f)
		}
		return xlist.ErrNotAvailable
	}
	return xlist.ErrReadOnlyMode
}

// Clear implements xlist.Writer interface
func (l *List) Clear(ctx context.Context) error {
	if !l.readonly {
		for _, child := range l.services {
			err := child.Clear(ctx)
			if err != nil {
				return err
			}
		}
	}
	return xlist.ErrReadOnlyMode
}

// ReadOnly implements xlist.Writer interface
func (l *List) ReadOnly() (bool, error) {
	return l.readonly, nil
}
