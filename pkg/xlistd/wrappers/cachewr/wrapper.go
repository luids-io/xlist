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

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// WrapperClass registered
const WrapperClass = "cache"

// DefaultConfig returns default configuration.
func DefaultConfig() Config {
	return Config{
		TTL:           defaultCacheTTL,
		Cleanups:      defaultCacheCleanups,
		RandomSeconds: defaultRandomCache,
		DoStats:       false,
	}
}

// Config options.
type Config struct {
	TTL             int
	NegativeTTL     int
	Cleanups        time.Duration
	RandomSeconds   int
	DoStats         bool
	ForceValidation bool
}

// Wrapper implements a cache for list checkers.
type Wrapper struct {
	cfg   Config
	list  xlistd.List
	cache *cacheimpl.Cache
}

var (
	defaultCacheTTL      = 300 //seconds
	defaultCacheCleanups = 6 * time.Minute
	defaultRandomCache   = 60 //seconds to randomize
)

// New returns a new wrapper.
func New(list xlistd.List, cfg Config) *Wrapper {
	//randomize cache cleanups
	rands := time.Duration(rand.Intn(cfg.RandomSeconds)) * time.Second
	c := &Wrapper{
		cfg:   cfg,
		cache: cacheimpl.New(time.Duration(cfg.TTL)*time.Second, cfg.Cleanups+rands),
		list:  list,
	}
	return c
}

// ID implements xlistd.List interface.
func (c *Wrapper) ID() string {
	return c.list.ID()
}

// Class implements xlistd.List interface.
func (c *Wrapper) Class() string {
	return c.list.Class()
}

// Check implements xlist.Checker interface.
func (c *Wrapper) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	name, ctx, err := xlist.DoValidation(ctx, name, resource, c.cfg.ForceValidation)
	if err != nil {
		return xlist.Response{}, err
	}
	resp, ok := c.get(name, resource)
	if ok {
		return resp, nil
	}
	resp, err = c.list.Check(ctx, name, resource)
	if err == nil {
		resp = c.set(name, resource, resp)
	}
	return resp, err
}

// Resources implements xlist.Checker interface.
func (c *Wrapper) Resources() []xlist.Resource {
	return c.list.Resources()
}

// Ping implements xlist.Checker interface.
func (c *Wrapper) Ping() error {
	return c.list.Ping()
}

// Flush deletes all items from cache.
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
	if (r.TTL == xlist.NeverCache) || (!r.Result && c.cfg.NegativeTTL == xlist.NeverCache) {
		return r
	}
	//sets cache
	ttl := c.cfg.TTL
	if !r.Result && c.cfg.NegativeTTL > 0 {
		ttl = c.cfg.NegativeTTL
	}
	if r.TTL < ttl { //minor than cachettl
		r.TTL = ttl //sets reponse to cachettl
	}
	key := fmt.Sprintf("%s_%s", resource.String(), name)
	c.cache.Set(key, r, time.Duration(r.TTL)*time.Second)
	return r
}
