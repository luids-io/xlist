// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package responsewr provides a wrapper for RBLs that changes the responses.
//
// This package is a work in progress and makes no API stability promises.
package responsewr

import (
	"context"
	"fmt"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/reason"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// Config options
type Config struct {
	Clean        bool
	Aggregate    bool
	Negate       bool
	Reason       string
	Preffix      string
	TTL          int
	UseThreshold bool
	Score        int
}

// Wrapper implements an xlist.Checker wrapper for change responses
type Wrapper struct {
	opts options
	list xlistd.List
}

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

// New returns a new Wrapper
func New(list xlistd.List, cfg Config) *Wrapper {
	if cfg.TTL < xlist.NeverCache {
		cfg.TTL = 0
	}
	return &Wrapper{
		opts: options{
			clean:        cfg.Clean,
			aggregate:    cfg.Aggregate,
			negate:       cfg.Negate,
			reason:       cfg.Reason,
			preffix:      cfg.Preffix,
			ttl:          cfg.TTL,
			useThreshold: cfg.UseThreshold,
			score:        cfg.Score,
		},
		list: list,
	}
}

// ID implements xlistd.List interface
func (w *Wrapper) ID() string {
	return w.list.ID()
}

// Class implements xlistd.List interface
func (w *Wrapper) Class() string {
	return BuildClass
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

// ReadOnly implements xlistd.List interface
func (w *Wrapper) ReadOnly() bool {
	return true
}
