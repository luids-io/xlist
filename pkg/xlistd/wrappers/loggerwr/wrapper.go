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

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/yalogi"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Rules: Rules{
			Found:    Info,
			NotFound: Debug,
			Error:    Warn,
		}}
}

// Config options
type Config struct {
	Prefix   string
	Rules    Rules
	ShowPeer bool
}

// Wrapper implements a logger for list checkers
type Wrapper struct {
	showPeer bool
	preffix  string
	log      Logger
	action   Rules
	list     xlistd.List
}

// Rules defines log levels for each event
type Rules struct {
	Found    LogLevel
	NotFound LogLevel
	Error    LogLevel
}

// New creates a logger wrapper with preffix
func New(list xlistd.List, logger yalogi.Logger, cfg Config) *Wrapper {
	return &Wrapper{
		showPeer: cfg.ShowPeer,
		preffix:  cfg.Prefix,
		log:      logger,
		action:   cfg.Rules,
		list:     list,
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
	//do check
	resp, err := w.list.Check(ctx, name, resource)

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
	peerInfo := w.getPeerInfo(ctx)
	//outputs event
	switch level {
	case Debug:
		w.log.Debugf("%s: [%s] Check('%s',%s) = %s (%s)", w.preffix, peerInfo, name, resource, result, resp.Reason)
	case Info:
		w.log.Infof("%s: [%s] Check('%s',%s) = %s (%s)", w.preffix, peerInfo, name, resource, result, resp.Reason)
	case Warn:
		w.log.Warnf("%s: [%s] Check('%s',%s) = %s (%s)", w.preffix, peerInfo, name, resource, result, resp.Reason)
	case Error:
		w.log.Errorf("%s: [%s] Check('%s',%s) = %s (%s)", w.preffix, peerInfo, name, resource, result, resp.Reason)
	}
	return resp, err
}

// Resources implements xlist.Checker interface
func (w *Wrapper) Resources() []xlist.Resource {
	res := w.list.Resources()
	w.log.Debugf("%s: Resources() = %v", w.preffix, res)
	return res
}

// Ping implements xlist.Checker interface
func (w *Wrapper) Ping() error {
	err := w.list.Ping()
	if err != nil {
		w.log.Warnf("%s: Ping() = %v", w.preffix, err)
		return err
	}
	w.log.Debugf("%s: Ping()", w.preffix)
	return nil
}

// ReadOnly implements xlist.Writer interface
func (w *Wrapper) ReadOnly() bool {
	return true
}

func (w *Wrapper) getPeerInfo(ctx context.Context) string {
	//get peer info
	peerInfo := ""
	if w.showPeer {
		p, ok := peer.FromContext(ctx)
		if ok {
			peerInfo = fmt.Sprintf("%v", p.Addr)
		}
	}
	return peerInfo
}
