// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"fmt"
	"time"

	"github.com/luisguillenc/yalogi"

	"github.com/luids-io/xlist/internal/config"
	"github.com/luids-io/xlist/pkg/builder"
	"github.com/luids-io/xlist/pkg/components/dnsxl"
)

//some defaults
var (
	DefaultTimeouts    = 1 * time.Second
	DefaultRateLimiter = "naive"
)

// Builder is a factory for an xlist builder
func Builder(cfg *config.BuilderCfg, logger yalogi.Logger) (*builder.Builder, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}
	b := builder.New(
		builder.SourcesDir(cfg.SourcesDir),
		builder.CertsDir(cfg.CertsDir),
		builder.SetLogger(logger))

	if !cfg.DNSxL.Empty() {
		err := setupDNSxL(cfg.DNSxL)
		if err != nil {
			return nil, err
		}
	}
	return b, nil
}

// setupDNSxL configures module
func setupDNSxL(cfg config.DNSxLCfg) error {
	err := cfg.Validate()
	if err != nil {
		return err
	}
	if cfg.UseResolvConf {
		resolver, err := dnsxl.NewResolverFromConf("/etc/resolv.conf")
		if err != nil {
			return fmt.Errorf("getting resolver: %v", err)
		}
		dnsxl.DefaultResolver(resolver)
	}
	if cfg.Resolvers == nil && len(cfg.Resolvers) > 0 {
		resolver, err := dnsxl.NewResolverRRPool(cfg.Resolvers)
		if err != nil {
			return fmt.Errorf("getting resolver: %v", err)
		}
		dnsxl.DefaultResolver(resolver)
	}
	// dnsclient
	dnsclient := &dnsxl.Client{}
	if cfg.TimeoutMSecs > 0 {
		dnsclient.Timeout = time.Duration(cfg.TimeoutMSecs) * time.Millisecond
	}
	builder.RegisterCheckerBuilder(dnsxl.BuildClass, dnsxl.Builder(dnsclient))
	return nil
}
