// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"fmt"
	"time"

	"github.com/luids-io/xlist/pkg/xlistd/components/sblookupxl"

	"github.com/luids-io/common/util"
	"github.com/luids-io/core/apiservice"
	"github.com/luids-io/core/yalogi"
	"github.com/luids-io/xlist/internal/config"
	"github.com/luids-io/xlist/pkg/xlistd"
	"github.com/luids-io/xlist/pkg/xlistd/components/dnsxl"
)

//some defaults
var (
	DefaultTimeouts    = 1 * time.Second
	DefaultRateLimiter = "naive"
)

// ListBuilder is a factory for an xlist builder
func ListBuilder(cfg *config.XListCfg, apisvc apiservice.Discover, logger yalogi.Logger) (*xlistd.Builder, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}
	b := xlistd.NewBuilder(apisvc,
		xlistd.DataDir(cfg.DataDir),
		xlistd.CertsDir(cfg.CertsDir),
		xlistd.SetLogger(logger),
	)
	return b, nil
}

// SetupDNSxL configures module
func SetupDNSxL(cfg *config.DNSxLCfg) error {
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
	dnsCfg := dnsxl.DefaultConfig()
	if cfg.TimeoutMSecs > 0 {
		dnsCfg.Timeout = time.Duration(cfg.TimeoutMSecs) * time.Millisecond
	}
	xlistd.RegisterListBuilder(dnsxl.ComponentClass, dnsxl.Builder(dnsCfg))
	return nil
}

// SetupSBLookup configures module
func SetupSBLookup(cfg *config.SBLookupCfg, logger yalogi.Logger) error {
	err := cfg.Validate()
	if err != nil {
		return err
	}
	sbCfg := sblookupxl.Config{}
	if cfg.APIKey != "" {
		sbCfg.APIKey = cfg.APIKey
	}
	if cfg.ServerURL != "" {
		sbCfg.ServerURL = cfg.ServerURL
	}
	logger.Infof("registrando builder con config: %v", sbCfg)
	xlistd.RegisterListBuilder(sblookupxl.ComponentClass, sblookupxl.Builder(sbCfg))
	return nil
}

//Lists creates lists from configuration files
func Lists(cfg *config.XListCfg, builder *xlistd.Builder, logger yalogi.Logger) error {
	err := cfg.Validate()
	if err != nil {
		return fmt.Errorf("bad config: %v", err)
	}
	dbfiles, err := util.GetFilesDB("json", cfg.ServiceFiles, cfg.ServiceDirs)
	if err != nil {
		return fmt.Errorf("loading dbfiles: %v", err)
	}
	defs, err := loadListDefs(dbfiles)
	if err != nil {
		return fmt.Errorf("loading dbfiles: %v", err)
	}
	for _, def := range defs {
		if def.Disabled {
			continue
		}
		_, err := builder.Build(def)
		if err != nil {
			return fmt.Errorf("creating '%s': %v", def.ID, err)
		}
	}
	return nil
}

func loadListDefs(dbFiles []string) ([]xlistd.ListDef, error) {
	loadedDB := make([]xlistd.ListDef, 0)
	for _, file := range dbFiles {
		entries, err := xlistd.ListDefsFromFile(file)
		if err != nil {
			return nil, fmt.Errorf("couln't load database: %v", err)
		}
		loadedDB = append(loadedDB, entries...)
	}
	return loadedDB, nil
}
