// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package geoip2xl provides a xlist.Checker implementation that uses a
// geoip database for checks. This means that the RBL can check if an ip
// is in a list of countries. It only allows IPv4 resources.
//
// This package is a work in progress and makes no API stability promises.
package geoip2xl

import (
	"context"
	"fmt"
	"net"
	"strings"

	geoip2 "github.com/oschwald/geoip2-golang"

	"github.com/luids-io/core/xlist"
)

// List implements an RBL that uses a geoip database for checks
type List struct {
	xlist.List
	opts     options
	started  bool
	rules    Rules
	dbPath   string
	database *geoip2.Reader
}

// Rules defines the logic for checks
type Rules struct {
	// Countries is a list of country codes
	Countries []string
	// Reverse the matching of the rule
	Reverse bool
}

// Option is used for configuration options
type Option func(*options)

type options struct {
	forceValidation bool
	reason          string
}

var defaultOptions = options{}

// ForceValidation forces components to ignore context and validate requests
func ForceValidation(b bool) Option {
	return func(o *options) {
		o.forceValidation = b
	}
}

// Reason sets a fixed reason for component responses
func Reason(s string) Option {
	return func(o *options) {
		o.reason = s
	}
}

// New constructs a new List with dbpath as database and rules for logic
func New(dbpath string, rules Rules, opt ...Option) *List {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	s := &List{
		opts:   opts,
		rules:  rules,
		dbPath: dbpath,
	}
	// capitalize country codes
	for idx, c := range s.rules.Countries {
		s.rules.Countries[idx] = strings.ToUpper(c)
	}
	return s
}

// Check implements xlist.Checker interface
func (l *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if !l.started {
		return xlist.Response{}, xlist.ErrNotAvailable
	}
	if resource != xlist.IPv4 {
		return xlist.Response{}, xlist.ErrNotImplemented
	}
	name, _, err := xlist.DoValidation(ctx, name, resource, l.opts.forceValidation)
	if err != nil {
		return xlist.Response{}, err
	}
	resp, err := l.checkRules(net.ParseIP(name))
	if err == nil && l.opts.reason != "" {
		resp.Reason = l.opts.reason
	}
	return resp, err
}

// Resources implements xlist.Checker interface
func (l *List) Resources() []xlist.Resource {
	return []xlist.Resource{xlist.IPv4}
}

// Ping implements xlist.Checker interface
func (l *List) Ping() error {
	if !l.started {
		return xlist.ErrNotAvailable
	}
	return nil
}

// Append implements xlist.Writer interface
func (l *List) Append(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	if !l.started {
		return xlist.ErrNotAvailable
	}
	return xlist.ErrReadOnlyMode
}

// Remove implements xlist.Writer interface
func (l *List) Remove(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	if !l.started {
		return xlist.ErrNotAvailable
	}
	return xlist.ErrReadOnlyMode
}

// Clear implements xlist.Writer interface
func (l *List) Clear(ctx context.Context) error {
	if !l.started {
		return xlist.ErrNotAvailable
	}
	return xlist.ErrReadOnlyMode
}

// ReadOnly implements xlist.Writer interface
func (l *List) ReadOnly() (bool, error) {
	if !l.started {
		return false, xlist.ErrNotAvailable
	}
	return true, nil
}

// Start opens database file
func (l *List) Start() error {
	var err error
	l.database, err = geoip2.Open(l.dbPath)
	if err == nil {
		l.started = true
	}
	return err
}

// Shutdown closes the database file
func (l *List) Shutdown() {
	if l.started {
		l.database.Close()
	}
}

func (l *List) checkRules(ip net.IP) (xlist.Response, error) {
	info, err := l.database.Country(ip)
	if err != nil {
		return xlist.Response{}, err
	}
	if info.Country.IsoCode == "" {
		return xlist.Response{}, fmt.Errorf("can't get code for ip %s", ip.String())
	}

	found := false
	reason := fmt.Sprintf("found country code '%s'", info.Country.IsoCode)
	for _, c := range l.rules.Countries {
		if c == info.Country.IsoCode {
			found = true
		}
	}
	if l.rules.Reverse {
		if found {
			found = false
		} else {
			found = true
		}
	}
	response := xlist.Response{}
	if found {
		response.Result = found
		response.Reason = reason
	}
	return response, nil
}
