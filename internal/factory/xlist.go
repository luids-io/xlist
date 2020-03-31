// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"fmt"
	"time"

	"github.com/luids-io/common/util"
	"github.com/luids-io/core/utils/yalogi"
	"github.com/luids-io/xlist/internal/config"
	"github.com/luids-io/xlist/pkg/components/dnsxl"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

//some defaults
var (
	DefaultTimeouts    = 1 * time.Second
	DefaultRateLimiter = "naive"
)

// ListBuilder is a factory for an xlist builder
func ListBuilder(cfg *config.XListCfg, logger yalogi.Logger) (*listbuilder.Builder, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}
	b := listbuilder.New(
		listbuilder.SourcesDir(cfg.SourcesDir),
		listbuilder.CertsDir(cfg.CertsDir),
		listbuilder.SetLogger(logger),
	)
	//modules default options
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
	listbuilder.RegisterListBuilder(dnsxl.BuildClass, dnsxl.Builder(dnsclient))
	return nil
}

//Lists creates lists from configuration files
func Lists(cfg *config.XListCfg, builder *listbuilder.Builder, logger yalogi.Logger) error {
	err := cfg.Validate()
	if err != nil {
		return fmt.Errorf("bad config: %v", err)
	}
	dbfiles, err := util.GetFilesDB("json", cfg.ConfigFiles, cfg.ConfigDirs)
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

func loadListDefs(dbFiles []string) ([]listbuilder.ListDef, error) {
	loadedDB := make([]listbuilder.ListDef, 0)
	for _, file := range dbFiles {
		entries, err := listbuilder.DefsFromFile(file)
		if err != nil {
			return nil, fmt.Errorf("couln't load database: %v", err)
		}
		loadedDB = append(loadedDB, entries...)
	}
	return loadedDB, nil
}
