// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package policywr provides a wrapper for RBLs that inserts policies into
// responses.
//
// This package is a work in progress and makes no API stability promises.
package policywr

import (
	"context"
	"fmt"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/reason"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// WrapperClass defines class name for component builder.
const WrapperClass = "policy"

// Config options
type Config struct {
	Merge        bool
	UseThreshold bool
	Score        int
}

// Wrapper implements an xlistd.List wrapper for insert policies.
type Wrapper struct {
	cfg    Config
	policy reason.Policy
	list   xlistd.List
}

// New returns a new Wrapper.
func New(list xlistd.List, p reason.Policy, cfg Config) *Wrapper {
	return &Wrapper{
		cfg:    cfg,
		policy: p,
		list:   list,
	}
}

// ID implements xlistd.List interface.
func (w *Wrapper) ID() string {
	return w.list.ID()
}

// Class implements xlistd.List interface.
func (w *Wrapper) Class() string {
	return w.list.Class()
}

// Check implements xlist.Checker interface.
func (w *Wrapper) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	resp, err := w.list.Check(ctx, name, resource)
	if err == nil && resp.Result {
		if w.cfg.UseThreshold {
			score, _, err := reason.ExtractScore(resp.Reason)
			if err != nil {
				return resp, err
			}
			if w.cfg.Score >= score {
				return resp, nil
			}
		}
		if w.cfg.Merge {
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

// Resources implements xlist.Checker interface.
func (w *Wrapper) Resources() []xlist.Resource {
	return w.list.Resources()
}

// Ping implements xlist.Checker interface.
func (w *Wrapper) Ping() error {
	return w.list.Ping()
}
