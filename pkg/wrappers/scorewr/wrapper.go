// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package scorewr provides a wrapper for RBLs that inserts scores into
// responses.
//
// This package is a work in progress and makes no API stability promises.
package scorewr

import (
	"context"
	"regexp"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/core/xlist/reason"
)

// Wrapper implements an xlist.Checker wrapper for insert scores
type Wrapper struct {
	opts       options
	score      int
	exprScores []exprScored
	checker    xlist.Checker
}

// Option is used for component configuration
type Option func(*options)

type options struct{}

var defaultOptions = options{}

type exprScored struct {
	expr  *regexp.Regexp
	score int
}

// New returns a new Wrapper
func New(list xlist.Checker, score int, opt ...Option) *Wrapper {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	return &Wrapper{
		opts:       opts,
		score:      score,
		exprScores: make([]exprScored, 0),
		checker:    list,
	}
}

// Check implements xlist.Checker interface
func (w *Wrapper) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	resp, err := w.checker.Check(ctx, name, resource)
	if err == nil && resp.Result {
		sumScore := 0
		matched := false
		for _, item := range w.exprScores {
			if item.expr.MatchString(resp.Reason) {
				matched = true
				sumScore = sumScore + item.score
			}
		}
		if matched {
			resp.Reason = reason.WithScore(sumScore, resp.Reason)
		} else {
			resp.Reason = reason.WithScore(w.score, resp.Reason)
		}
	}
	return resp, err
}

// AddExpr adds expresion score
func (w *Wrapper) AddExpr(expr string, score int) error {
	r, err := regexp.Compile(expr)
	if err != nil {
		return err
	}
	w.exprScores = append(w.exprScores, exprScored{expr: r, score: score})
	return nil
}

// Resources implements xlist.Checker interface
func (w *Wrapper) Resources() []xlist.Resource {
	return w.checker.Resources()
}

// Ping implements xlist.Checker interface
func (w *Wrapper) Ping() error {
	return w.checker.Ping()
}
