// Copyright 2020 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package sblookupxl provides a xlistd.List implementation that uses
// google safe browsing api as source.
//
// This package is a work in progress and makes no API stability promises.
package sblookupxl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/safebrowsing"

	"github.com/luids-io/api/xlist"
)

// ComponentClass registered.
const ComponentClass = "sblookup"

// Config options.
type Config struct {
	ServerURL       string
	APIKey          string
	Database        string
	ForceValidation bool
	Reason          string
	Threats         []string
}

type options struct {
	forceValidation bool
	reason          string
}

// List implements an RBL that uses google safe browsing api as source
type List struct {
	id       string
	opts     options
	closed   bool
	sbrowser *safebrowsing.SafeBrowser
}

// New constructs a new List with config.
func New(id string, cfg Config) (*List, error) {
	threats, err := getThreats(cfg.Threats)
	if err != nil {
		return nil, err
	}
	sb, err := safebrowsing.NewSafeBrowser(safebrowsing.Config{
		ServerURL:   cfg.ServerURL,
		APIKey:      cfg.APIKey,
		DBPath:      cfg.Database,
		ThreatLists: threats,
	})
	if err != nil {
		return nil, err
	}
	l := &List{
		id:       id,
		sbrowser: sb,
		opts: options{
			forceValidation: cfg.ForceValidation,
			reason:          cfg.Reason,
		},
	}
	return l, nil
}

// ID implements xlistd.List interface.
func (l *List) ID() string {
	return l.id
}

// Class implements xlistd.List interface.
func (l *List) Class() string {
	return ComponentClass
}

// Check implements xlist.Checker interface.
func (l *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if l.closed {
		return xlist.Response{}, xlist.ErrUnavailable
	}
	if resource != xlist.Domain {
		return xlist.Response{}, xlist.ErrNotSupported
	}
	name, _, err := xlist.DoValidation(ctx, name, resource, l.opts.forceValidation)
	if err != nil {
		return xlist.Response{}, err
	}
	threads, err := l.sbrowser.LookupURLsContext(ctx, []string{name})
	if err != nil {
		return xlist.Response{}, err
	}
	if len(threads[0]) > 0 {
		if l.opts.reason != "" {
			return xlist.Response{Result: true, Reason: l.opts.reason}, nil
		}
		reasons := make([]string, 0, len(threads[0]))
		for _, t := range threads[0] {
			reasons = append(reasons, fmt.Sprintf("%v", t.ThreatDescriptor))
		}
		return xlist.Response{Result: true, Reason: strings.Join(reasons, ",")}, nil
	}
	return xlist.Response{}, nil
}

// Resources implements xlist.Checker interface.
func (l *List) Resources(ctx context.Context) ([]xlist.Resource, error) {
	return []xlist.Resource{xlist.Domain}, nil
}

// Ping implements xlistd.List interface.
func (l *List) Ping() error {
	if l.closed {
		return errors.New("list is closed")
	}
	_, err := l.sbrowser.Status()
	return err
}

// Close closes the database file.
func (l *List) Close() {
	if !l.closed {
		l.closed = true
		l.sbrowser.Close()
	}
}

func getThreats(items []string) ([]safebrowsing.ThreatDescriptor, error) {
	var ret []safebrowsing.ThreatDescriptor
	if len(items) == 0 {
		//ret nil slice
		return ret, nil
	}
	ret = make([]safebrowsing.ThreatDescriptor, 0, len(items))
	for _, i := range items {
		ttype, err := getThreatType(i)
		if err != nil {
			return nil, err
		}
		ret = append(ret, safebrowsing.ThreatDescriptor{
			ThreatType:      ttype,
			ThreatEntryType: safebrowsing.ThreatEntryType_URL,
			PlatformType:    safebrowsing.PlatformType_AllPlatforms,
		})
	}
	return ret, nil
}

func getThreatType(s string) (safebrowsing.ThreatType, error) {
	s = strings.ToLower(s)
	switch s {
	case "malware":
		return safebrowsing.ThreatType_Malware, nil
	case "phishing":
		return safebrowsing.ThreatType_SocialEngineering, nil
	case "social_engineering":
		return safebrowsing.ThreatType_SocialEngineering, nil
	case "unwanted":
		return safebrowsing.ThreatType_UnwantedSoftware, nil
	case "unwanted_software":
		return safebrowsing.ThreatType_UnwantedSoftware, nil
	}
	return safebrowsing.ThreatType_Malware, fmt.Errorf("unknown ThreatType value '%s'", s)
}
