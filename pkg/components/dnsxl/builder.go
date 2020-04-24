// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package dnsxl

import (
	"errors"
	"fmt"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

// BuildClass defines default class for component builder
const BuildClass = "dnsxl"

// Builder returns a builder for component dnsxl
func Builder(cfg Config) listbuilder.BuildListFn {
	return func(builder *listbuilder.Builder, parents []string, def listbuilder.ListDef) (xlist.List, error) {
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
		// create dnsxl object
		return New(def.Source, cfg)
	}
}

func newResolversFromDef(def listbuilder.ListDef) (Resolver, error) {
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

func parseOptions(cfg Config, opts map[string]interface{}) (Config, error) {
	rCfg := cfg
	pingdns, ok, err := option.String(opts, "pingdns")
	if err != nil {
		return rCfg, err
	}
	if ok {
		if !isDomain(pingdns) {
			return rCfg, errors.New("invalid 'pingdns'")
		}
		rCfg.PingDNS = pingdns
	}

	halfping, ok, err := option.Bool(opts, "halfping")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.HalfPing = halfping
	}

	authtoken, ok, err := option.String(opts, "authtoken")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.AuthToken = authtoken
	}

	resolvreason, ok, err := option.Bool(opts, "resolvreason")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.ResolveReason = resolvreason
	}

	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.Reason = reason
	}

	retries, ok, err := option.Int(opts, "retries")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.Retries = retries
	}

	dnscodes, ok, err := option.HashString(opts, "dnscodes")
	if err != nil {
		return rCfg, err
	}
	if ok {
		cfg.DNSCodes = make(map[string]string, len(dnscodes))
		for k, v := range dnscodes {
			if !isIP(k) {
				return rCfg, fmt.Errorf("invalid ip address '%s' in dnscodes", k)
			}
			cfg.DNSCodes[k] = v
		}
	}

	errcodes, ok, err := option.HashString(opts, "errcodes")
	if err != nil {
		return rCfg, err
	}
	if ok {
		cfg.ErrCodes = make(map[string]string, len(errcodes))
		for k, v := range errcodes {
			if !isIP(k) {
				return rCfg, fmt.Errorf("invalid ip address '%s' in errcodes", k)
			}
			cfg.ErrCodes[k] = v
		}
	}

	return rCfg, nil
}

func init() {
	listbuilder.RegisterListBuilder(BuildClass, Builder(DefaultConfig()))
}
