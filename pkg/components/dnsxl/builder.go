// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package dnsxl

import (
	"errors"
	"fmt"

	"github.com/miekg/dns"

	"github.com/luids-io/core/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

// BuildClass defines default class for component builder
const BuildClass = "dnsxl"

// Builder returns a builder for component dnsxl
func Builder(client *dns.Client, opt ...Option) listbuilder.BuildListFn {
	return func(builder *listbuilder.Builder, parents []string, def listbuilder.ListDef) (xlist.List, error) {
		if def.Source == "" {
			def.Source = def.ID
		}
		bopt := make([]Option, 0)
		bopt = append(bopt, opt...)
		if def.Opts != nil {
			var err error
			bopt, err = parseOptions(bopt, def.Opts)
			if err != nil {
				return nil, err
			}
			//custom resolver
			_, ok1 := def.Opts["resolvers"]
			_, ok2 := def.Opts["nsresolvers"]
			if ok1 || ok2 {
				pool, err := newResolversFromDef(client, def)
				if err != nil {
					return nil, fmt.Errorf("creating resolvers: %v", err)
				}
				bopt = append(bopt, UseResolver(pool))
			}
		}

		// create dnsxl object
		bl := New(client, def.Source, def.Resources, bopt...)
		if def.Opts != nil {
			dnscodes, _, err := option.HashString(def.Opts, "dnscodes")
			if err != nil {
				return nil, err
			}
			for k, v := range dnscodes {
				bl.AddDNSCode(k, v)
			}
			errcodes, _, err := option.HashString(def.Opts, "errcodes")
			if err != nil {
				return nil, err
			}
			for k, v := range errcodes {
				bl.AddErrCode(k, v)
			}
		}
		return bl, nil
	}
}

func newResolversFromDef(client *dns.Client, def listbuilder.ListDef) (Resolver, error) {
	usens, _, err := option.Bool(def.Opts, "nsresolvers")
	if err != nil {
		return nil, err
	}
	if usens {
		pool, err := NewResolverPoolFromZone(client, def.Source)
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

func parseOptions(bopt []Option, opts map[string]interface{}) ([]Option, error) {
	pingdns, ok, err := option.String(opts, "pingdns")
	if err != nil {
		return bopt, err
	}
	if ok {
		if !isDomain(pingdns) {
			return bopt, errors.New("invalid 'pingdns'")
		}
		bopt = append(bopt, UseDNSPing(pingdns))
	}

	halfping, ok, err := option.Bool(opts, "halfping")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, HalfPing(halfping))
	}

	authtoken, ok, err := option.String(opts, "authtoken")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, AuthToken(authtoken))
	}

	resolvreason, ok, err := option.Bool(opts, "resolvreason")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, ResolveReason(resolvreason))
	}

	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, Reason(reason))
	}

	retries, ok, err := option.Int(opts, "retries")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, Retries(retries))
	}
	return bopt, nil
}

func init() {
	listbuilder.RegisterListBuilder(BuildClass, Builder(&dns.Client{}))
}
