// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package policywr provides a wrapper for RBLs that inserts policies into
// responses.
//
// This package is a work in progress and makes no API stability promises.
package policywr

import (
	"context"
	"fmt"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/core/xlist/reason"
)

// Config options
type Config struct {
	Merge        bool
	UseThreshold bool
	Score        int
}

// Wrapper implements an xlist.List wrapper for insert policies
type Wrapper struct {
	opts   options
	policy reason.Policy
	list   xlist.List
}

type options struct {
	merge        bool
	useThreshold bool
	score        int
}

// New returns a new Wrapper
func New(list xlist.List, p reason.Policy, cfg Config) *Wrapper {
	return &Wrapper{
		opts: options{
			merge:        cfg.Merge,
			useThreshold: cfg.UseThreshold,
			score:        cfg.Score,
		},
		policy: p,
		list:   list,
	}
}

// Check implements xlist.Checker interface
func (w *Wrapper) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	resp, err := w.list.Check(ctx, name, resource)
	if err == nil && resp.Result {
		if w.opts.useThreshold {
			score, _, err := reason.ExtractScore(resp.Reason)
			if err != nil {
				return resp, err
			}
			if w.opts.score >= score {
				return resp, nil
			}
		}
		if w.opts.merge {
			p, r, merr := reason.ExtractPolicy(resp.Reason)
			if merr != nil {
				//do nothing
				return resp, err
			}
			p.Merge(w.policy)
			resp.Reason = fmt.Sprintf("%s%s", p.String(), r)
		} else {
			resp.Reason = reason.WithPolicy(w.policy, resp.Reason)
		}
	}
	return resp, err
}

// Resources implements xlist.Checker interface
func (w *Wrapper) Resources() []xlist.Resource {
	return w.list.Resources()
}

// Ping implements xlist.Checker interface
func (w *Wrapper) Ping() error {
	return w.list.Ping()
}

// Append implements xlist.Writer interface
func (w *Wrapper) Append(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return w.list.Append(ctx, name, r, f)
}

// Remove implements xlist.Writer interface
func (w *Wrapper) Remove(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return w.list.Remove(ctx, name, r, f)
}

// Clear implements xlist.Writer interface
func (w *Wrapper) Clear(ctx context.Context) error {
	return w.list.Clear(ctx)
}

// ReadOnly implements xlist.Writer interface
func (w *Wrapper) ReadOnly() (bool, error) {
	return w.list.ReadOnly()
}
