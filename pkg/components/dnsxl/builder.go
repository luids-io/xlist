// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package dnsxl

import (
	"errors"
	"fmt"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines default class for component builder
const BuildClass = "dnsxl"

// Builder returns a builder for component dnsxl
func Builder(defaultCfg Config) builder.BuildListFn {
	return func(b *builder.Builder, parents []string, def builder.ListDef) (xlist.List, error) {
		cfg := defaultCfg.Copy()
		if def.Source == "" {
			def.Source = def.ID
		}
		if def.Opts != nil {
			var err error
			cfg, err = parseOptions(cfg, def.Opts)
			if err != nil {
				return nil, err
			}
			//custom resolver
			_, ok1 := def.Opts["resolvers"]
			_, ok2 := def.Opts["nsresolvers"]
			if ok1 || ok2 {
				pool, err := newResolversFromDef(def)
				if err != nil {
					return nil, fmt.Errorf("creating resolvers: %v", err)
				}
				cfg.Resolver = pool
			}
		}
		return New(def.ID, def.Source, def.Resources, cfg)
	}
}

func newResolversFromDef(def builder.ListDef) (Resolver, error) {
	usens, _, err := option.Bool(def.Opts, "nsresolvers")
	if err != nil {
		return nil, err
	}
	if usens {
		pool, err := NewResolverPoolFromZone(def.Source)
		if err != nil {
			return nil, fmt.Errorf("can't use zone nameservers: %v", err)
		}
		return pool, nil
	}

	resolvers, ok, err := option.SliceString(def.Opts, "resolvers")
	if err != nil {
		return nil, err
	}
	if ok {
		pool, err := NewResolverRRPool(resolvers)
		if err != nil {
			return nil, fmt.Errorf("can't use resolvers: %v", err)
		}
		return pool, nil
	}

	return nil, errors.New("no valid resolver options")
}

func parseOptions(src Config, opts map[string]interface{}) (Config, error) {
	dst := src
	pingdns, ok, err := option.String(opts, "pingdns")
	if err != nil {
		return dst, err
	}
	if ok {
		if !isDomain(pingdns) {
			return dst, errors.New("invalid 'pingdns'")
		}
		dst.PingDNS = pingdns
	}

	halfping, ok, err := option.Bool(opts, "halfping")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.HalfPing = halfping
	}

	authtoken, ok, err := option.String(opts, "authtoken")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.AuthToken = authtoken
	}

	resolvreason, ok, err := option.Bool(opts, "resolvreason")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.ResolveReason = resolvreason
	}

	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Reason = reason
	}

	retries, ok, err := option.Int(opts, "retries")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Retries = retries
	}

	dnscodes, ok, err := option.HashString(opts, "dnscodes")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.DNSCodes = make(map[string]string, len(dnscodes))
		for k, v := range dnscodes {
			if !isIP(k) {
				return dst, fmt.Errorf("invalid ip address '%s' in dnscodes", k)
			}
			dst.DNSCodes[k] = v
		}
	}

	errcodes, ok, err := option.HashString(opts, "errcodes")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.ErrCodes = make(map[string]string, len(errcodes))
		for k, v := range errcodes {
			if !isIP(k) {
				return dst, fmt.Errorf("invalid ip address '%s' in errcodes", k)
			}
			dst.ErrCodes[k] = v
		}
	}
	return dst, nil
}

func init() {
	builder.RegisterListBuilder(BuildClass, Builder(DefaultConfig()))
}
