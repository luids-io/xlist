// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package cachewr provides a wrapper for RBLs that implements a memory cache
// system.
//
// This package is a work in progress and makes no API stability promises.
package cachewr

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	cacheimpl "github.com/patrickmn/go-cache"

	"github.com/luids-io/core/xlist"
)

// Wrapper implements a cache for list checkers
type Wrapper struct {
	opts    options
	checker xlist.Checker
	cache   *cacheimpl.Cache
}

// Option is used for component configuration
type Option func(*options)

type options struct {
	ttl             int
	negativettl     int
	cleanups        time.Duration
	randomSeconds   int
	doStats         bool
	forceValidation bool
}

var (
	defaultCacheTTL      = 300 //seconds
	defaultCacheCleanups = 6 * time.Minute
	defaultRandomCache   = 60 //seconds to randomize
)

var defaultOptions = options{
	ttl:           defaultCacheTTL,
	cleanups:      defaultCacheCleanups,
	randomSeconds: defaultRandomCache,
	doStats:       false,
}

// Cleanups sets time for cache cleanups
func Cleanups(d time.Duration, randomSeconds int) Option {
	return func(o *options) {
		o.cleanups = d
		o.randomSeconds = randomSeconds
	}
}

// TTL sets time cache in seconds
func TTL(ttl int) Option {
	return func(o *options) {
		if ttl > 0 {
			o.ttl = ttl
		}
	}
}

// NegativeTTL sets time for negative cache in seconds
func NegativeTTL(ttl int) Option {
	return func(o *options) {
		o.negativettl = ttl
	}
}

// ForceValidation forces components to ignore context and validate requests
func ForceValidation(b bool) Option {
	return func(o *options) {
		o.forceValidation = b
	}
}

// New returns a new wrapper
func New(checker xlist.Checker, opt ...Option) *Wrapper {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	//randomize cache cleanups
	rands := time.Duration(rand.Intn(opts.randomSeconds)) * time.Second
	c := &Wrapper{
		opts:    opts,
		cache:   cacheimpl.New(time.Duration(opts.ttl)*time.Second, opts.cleanups+rands),
		checker: checker,
	}
	return c
}

// Check implements xlist.Checker interface
func (c *Wrapper) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	name, ctx, err := xlist.DoValidation(ctx, name, resource, c.opts.forceValidation)
	if err != nil {
		return xlist.Response{}, err
	}
	resp, ok := c.get(name, resource)
	if ok {
		return resp, nil
	}
	resp, err = c.checker.Check(ctx, name, resource)
	if err == nil {
		resp = c.set(name, resource, resp)
	}
	return resp, err
}

// Resources implements xlist.Checker interface
func (c *Wrapper) Resources() []xlist.Resource {
	return c.checker.Resources()
}

// Ping implements xlist.Checker interface
func (c *Wrapper) Ping() error {
	return c.checker.Ping()
}

// Flush deletes all items from cache
func (c *Wrapper) Flush() {
	c.cache.Flush()
}

func (c *Wrapper) get(name string, resource xlist.Resource) (xlist.Response, bool) {
	key := fmt.Sprintf("%s_%s", resource.String(), name)
	hit, exp, ok := c.cache.GetWithExpiration(key)
	if ok {
		r := hit.(xlist.Response)
		if r.TTL >= 0 {
			//updates ttl
			ttl := exp.Sub(time.Now()).Seconds()
			if ttl < 0 { //nonsense
				panic("cache missfunction")
			}
			r.TTL = int(ttl)
		}
		return r, true
	}
	return xlist.Response{}, false
}

func (c *Wrapper) set(name string, resource xlist.Resource, r xlist.Response) xlist.Response {
	//if don't cache
	if (r.TTL == xlist.NeverCache) || (!r.Result && c.opts.negativettl == xlist.NeverCache) {
		return r
	}
	//sets cache
	ttl := c.opts.ttl
	if !r.Result && c.opts.negativettl > 0 {
		ttl = c.opts.negativettl
	}
	if r.TTL < ttl { //minor than cachettl
		r.TTL = ttl //sets reponse to cachettl
	}
	key := fmt.Sprintf("%s_%s", resource.String(), name)
	c.cache.Set(key, r, time.Duration(r.TTL)*time.Second)
	return r
}
