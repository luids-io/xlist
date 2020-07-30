// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package scorewr provides a wrapper for RBLs that inserts scores into
// responses.
//
// This package is a work in progress and makes no API stability promises.
package scorewr

import (
	"context"
	"regexp"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/reason"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// Config options
type Config struct {
	Scores []ScoreExpr
}

// ScoreExpr defines score matching
type ScoreExpr struct {
	RegExp *regexp.Regexp
	Score  int
}

// Match returns true if matching
func (e ScoreExpr) Match(s string) bool {
	if e.RegExp != nil && e.RegExp.MatchString(s) {
		return true
	}
	return false
}

// Wrapper implements an xlist.Checker wrapper for insert scores
type Wrapper struct {
	score      int
	exprScores []ScoreExpr
	list       xlistd.List
}

// New returns a new Wrapper
func New(list xlistd.List, score int, cfg Config) *Wrapper {
	w := &Wrapper{
		score: score,
		list:  list,
	}
	if len(cfg.Scores) > 0 {
		w.exprScores = make([]ScoreExpr, len(cfg.Scores), len(cfg.Scores))
		copy(w.exprScores, cfg.Scores)
	}
	return w
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
	if err == nil && resp.Result {
		sumScore := 0
		matched := false
		for _, item := range w.exprScores {
			if item.Match(resp.Reason) {
				matched = true
				sumScore = sumScore + item.Score
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
