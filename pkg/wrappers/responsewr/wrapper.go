// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package responsewr provides a wrapper for RBLs that changes the responses.
//
// This package is a work in progress and makes no API stability promises.
package responsewr

import (
	"context"
	"fmt"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/core/xlist/reason"
)

// Wrapper implements an xlist.Checker wrapper for change responses
type Wrapper struct {
	xlist.List

	opts options
	list xlist.List
}

// Option is used for component configuration
type Option func(*options)

type options struct {
	clean        bool
	aggregate    bool
	negate       bool
	reason       string
	preffix      string
	ttl          int
	useThreshold bool
	score        int
}

var defaultOptions = options{}

// Clean the response reason
func Clean(b bool) Option {
	return func(o *options) {
		o.clean = b
	}
}

// Aggregate encoded data
func Aggregate(b bool) Option {
	return func(o *options) {
		o.aggregate = b
	}
}

// Negate changes the response result
func Negate(b bool) Option {
	return func(o *options) {
		o.negate = b
	}
}

// Reason overrides the reason in case of positive result
func Reason(s string) Option {
	return func(o *options) {
		o.reason = s
	}
}

// PreffixReason adds a prefix to reason in case of positive result
func PreffixReason(s string) Option {
	return func(o *options) {
		o.preffix = s
	}
}

// TTL changes the ttl of the response to a fixed value
func TTL(i int) Option {
	return func(o *options) {
		if i >= xlist.NeverCache {
			o.ttl = i
		}
	}
}

// Threshold sets limit score for true reasons
func Threshold(i int) Option {
	return func(o *options) {
		o.useThreshold = true
		o.score = i
	}
}

// New returns a new Wrapper
func New(list xlist.List, opt ...Option) *Wrapper {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	return &Wrapper{
		opts: opts,
		list: list,
	}
}

// Check implements xlist.Checker interface
func (w *Wrapper) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	resp, err := w.list.Check(ctx, name, resource)
	if err == nil {
		if resp.Result {
			score, rest, err := reason.ExtractScore(resp.Reason)
			if err != nil {
				return resp, err
			}
			if w.opts.aggregate {
				resp.Reason = reason.WithScore(score, rest)
			}
			if w.opts.useThreshold && w.opts.score >= score {
				resp.Result = false
				resp.Reason = ""
			}
		}
		if w.opts.negate {
			if resp.Result {
				resp.Result = false
				resp.Reason = ""
			} else {
				resp.Result = true
			}
		}
		if resp.Result && w.opts.reason != "" {
			resp.Reason = w.opts.reason
		}
		if resp.Result && w.opts.preffix != "" {
			resp.Reason = fmt.Sprintf("%s: %s", w.opts.preffix, resp.Reason)
		}
		if resp.Result && w.opts.clean {
			resp.Reason = reason.Clean(resp.Reason)
		}
		if w.opts.ttl != 0 {
			resp.TTL = w.opts.ttl
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
