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

// Config options
type Config struct {
	Database string
	// Countries is a list of country codes
	Countries []string
	// Reverse the matching of the rule
	Reverse bool
	// Common options
	ForceValidation bool
	Reason          string
}

// Copy configuration
func (src Config) Copy() Config {
	dst := src
	if len(src.Countries) > 0 {
		dst.Countries = make([]string, len(src.Countries), len(src.Countries))
		copy(dst.Countries, src.Countries)
	}
	return dst
}

type options struct {
	forceValidation bool
	reason          string
}

// List implements an RBL that uses a geoip database for checks
type List struct {
	opts      options
	started   bool
	dbPath    string
	database  *geoip2.Reader
	countries []string
	reverse   bool
}

// New constructs a new List with dbpath as database and rules for logic
func New(database string, cfg Config) *List {
	l := &List{
		dbPath: database,
		opts: options{
			forceValidation: cfg.ForceValidation,
			reason:          cfg.Reason,
		},
		countries: make([]string, 0, len(cfg.Countries)),
		reverse:   cfg.Reverse,
	}
	for _, c := range cfg.Countries {
		l.countries = append(l.countries, strings.ToUpper(c))
	}
	return l
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
		return true, xlist.ErrNotAvailable
	}
	return true, nil
}

// Open opens database file
func (l *List) Open() error {
	var err error
	l.database, err = geoip2.Open(l.dbPath)
	if err == nil {
		l.started = true
	}
	return err
}

// Close closes the database file
func (l *List) Close() {
	if l.started {
		l.started = false
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
	found := true
	reason := fmt.Sprintf("found country code '%s'", info.Country.IsoCode)
	if len(l.countries) > 0 {
		found = false
		for _, c := range l.countries {
			if c == info.Country.IsoCode {
				found = true
			}
		}
		if l.reverse {
			if found {
				found = false
			} else {
				found = true
			}
		}
	}
	response := xlist.Response{}
	if found {
		response.Result = found
		response.Reason = reason
	}
	return response, nil
}
