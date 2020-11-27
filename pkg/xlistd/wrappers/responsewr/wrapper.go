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

// WrapperClass registered.
const WrapperClass = "response"

// Config options.
type Config struct {
	Clean        bool
	Aggregate    bool
	Negate       bool
	Reason       string
	Preffix      string
	TTL          int
	NegativeTTL  int
	UseThreshold bool
	Score        int
}

// Wrapper implements an xlistd.List wrapper for change responses.
type Wrapper struct {
	cfg  Config
	list xlistd.List
}

// New returns a new Wrapper.
func New(list xlistd.List, cfg Config) *Wrapper {
	if cfg.TTL < xlist.NeverCache {
		cfg.TTL = 0
	}
	if cfg.NegativeTTL < xlist.NeverCache {
		cfg.NegativeTTL = 0
	}
	return &Wrapper{cfg: cfg, list: list}
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
	if err == nil {
		if resp.Result {
			score, rest, err := reason.ExtractScore(resp.Reason)
			if err != nil {
				return resp, err
			}
			if w.cfg.Aggregate {
				resp.Reason = reason.WithScore(score, rest)
			}
			if w.cfg.UseThreshold && w.cfg.Score >= score {
				resp.Result = false
				resp.Reason = ""
			}
		}
		if w.cfg.Negate {
			if resp.Result {
				resp.Result = false
				resp.Reason = ""
			} else {
				resp.Result = true
			}
		}
		if resp.Result && w.cfg.Reason != "" {
			resp.Reason = w.cfg.Reason
		}
		if resp.Result && w.cfg.Preffix != "" {
			resp.Reason = fmt.Sprintf("%s: %s", w.cfg.Preffix, resp.Reason)
		}
		if resp.Result && w.cfg.Clean {
			resp.Reason = reason.Clean(resp.Reason)
		}
		if resp.Result && w.cfg.TTL != 0 {
			resp.TTL = w.cfg.TTL
		}
		if !resp.Result && w.cfg.NegativeTTL != 0 {
			resp.TTL = w.cfg.NegativeTTL
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
