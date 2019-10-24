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

// Wrapper implements an xlist.Checker wrapper for insert policies
type Wrapper struct {
	opts    options
	policy  reason.Policy
	checker xlist.Checker
}

// Option is used for component configuration
type Option func(*options)

type options struct {
	merge        bool
	useThreshold bool
	score        int
}

var defaultOptions = options{}

// Merge option merges response with the policy
func Merge(b bool) Option {
	return func(o *options) {
		o.merge = b
	}
}

// Threshold sets limit score for apply policy
func Threshold(i int) Option {
	return func(o *options) {
		o.useThreshold = true
		o.score = i
	}
}

// New returns a new Wrapper
func New(list xlist.Checker, p reason.Policy, opt ...Option) *Wrapper {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	return &Wrapper{
		opts:    opts,
		policy:  p,
		checker: list,
	}
}

// Check implements xlist.Checker interface
func (w *Wrapper) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	resp, err := w.checker.Check(ctx, name, resource)
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
	return w.checker.Resources()
}

// Ping implements xlist.Checker interface
func (w *Wrapper) Ping() error {
	return w.checker.Ping()
}
