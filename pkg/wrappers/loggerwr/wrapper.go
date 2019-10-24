// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package loggerwr provides a wrapper for RBLs that implements a logging
// system.
//
// This package is a work in progress and makes no API stability promises.
package loggerwr

import (
	"context"
	"fmt"

	"google.golang.org/grpc/peer"

	"github.com/luids-io/core/xlist"
)

// Wrapper implements a logger for list checkers
type Wrapper struct {
	opts    options
	preffix string
	log     Logger
	action  Rules
	checker xlist.Checker
}

// Option is used for component configuration
type Option func(*options)

var defaultOptions = options{}

type options struct {
	showPeer bool
}

// DefaultRules returns a new ruleset with default values
func DefaultRules() Rules {
	return Rules{
		Found:    Info,
		NotFound: Debug,
		Error:    Warn,
	}
}

// Rules defines log levels for each event
type Rules struct {
	Found    LogLevel
	NotFound LogLevel
	Error    LogLevel
}

// ShowPeer enables log peer context information
func ShowPeer(b bool) Option {
	return func(o *options) {
		o.showPeer = b
	}
}

// New creates a logger wrapper with preffix
func New(preffix string, checker xlist.Checker, log Logger, rules Rules, opt ...Option) *Wrapper {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	return &Wrapper{
		opts:    opts,
		preffix: preffix,
		log:     log,
		action:  rules,
		checker: checker,
	}
}

// Check implements xlist.Checker interface
func (w *Wrapper) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	//do check
	resp, err := w.checker.Check(ctx, name, resource)

	//get level
	var level LogLevel
	var result string
	if err != nil {
		level = w.action.Error
		result = fmt.Sprintf("error %v", err)
	} else {
		if resp.Result {
			level = w.action.Found
			result = "positive"
		} else {
			level = w.action.NotFound
			result = "negative"
		}
	}
	//get peer info
	peerInfo := ""
	if w.opts.showPeer {
		p, ok := peer.FromContext(ctx)
		if ok {
			peerInfo = fmt.Sprintf("%v", p.Addr)
		}
	}
	//outputs event
	switch level {
	case Debug:
		if w.opts.showPeer {
			w.log.Debugf("%s: %s check('%s',%s) = %s (%s)", w.preffix, peerInfo, name, resource, result, resp.Reason)
		} else {
			w.log.Debugf("%s: check('%s',%s) = %s (%s)", w.preffix, name, resource, result, resp.Reason)
		}
	case Info:
		if w.opts.showPeer {
			w.log.Infof("%s: %s check('%s',%s) = %s (%s)", w.preffix, peerInfo, name, resource, result, resp.Reason)
		} else {
			w.log.Infof("%s: check('%s',%s) = %s (%s)", w.preffix, name, resource, result, resp.Reason)
		}
	case Warn:
		if w.opts.showPeer {
			w.log.Warnf("%s: %s check('%s',%s) = %s (%s)", w.preffix, peerInfo, name, resource, result, resp.Reason)
		} else {
			w.log.Warnf("%s: check('%s',%s) = %s (%s)", w.preffix, name, resource, result, resp.Reason)
		}
	case Error:
		if w.opts.showPeer {
			w.log.Errorf("%s: %s check('%s',%s) = %s (%s)", w.preffix, peerInfo, name, resource, result, resp.Reason)
		} else {
			w.log.Errorf("%s: check('%s',%s) = %s (%s)", w.preffix, name, resource, result, resp.Reason)
		}
	}
	return resp, err
}

// Resources implements xlist.Checker interface
func (w *Wrapper) Resources() []xlist.Resource {
	res := w.checker.Resources()
	w.log.Debugf("%s: resources() = %v", w.preffix, res)
	return res
}

// Ping implements xlist.Checker interface
func (w *Wrapper) Ping() error {
	err := w.checker.Ping()
	if err != nil {
		w.log.Debugf("%s: ping() = %v", w.preffix, err)
	} else {
		w.log.Debugf("%s: ping()", w.preffix)
	}
	return err
}
