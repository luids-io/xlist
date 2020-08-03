// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/luids-io/common/util"
)

// XListCfg stores lists config paths and builder prefs
type XListCfg struct {
	//service configuration
	ServiceDirs  []string
	ServiceFiles []string
	//generic build opts
	DataDir  string
	CertsDir string
}

// SetPFlags setups posix flags for commandline configuration
func (cfg *XListCfg) SetPFlags(short bool, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	if short {
		pflag.StringSliceVarP(&cfg.ServiceDirs, aprefix+"service.dirs", "D", cfg.ServiceDirs, "Configuration service dirs.")
		pflag.StringSliceVarP(&cfg.ServiceFiles, aprefix+"service.files", "d", cfg.ServiceFiles, "Configuration service files.")
	} else {
		pflag.StringSliceVar(&cfg.ServiceDirs, aprefix+"service.dirs", cfg.ServiceDirs, "Configuration service dirs.")
		pflag.StringSliceVar(&cfg.ServiceFiles, aprefix+"service.files", cfg.ServiceFiles, "Configuration service files.")
	}
	pflag.StringVar(&cfg.DataDir, aprefix+"datadir", cfg.DataDir, "Path to data files.")
	pflag.StringVar(&cfg.CertsDir, aprefix+"certsdir", cfg.CertsDir, "Path to certificate files.")
}

// BindViper setups posix flags for commandline configuration and bind to viper
func (cfg *XListCfg) BindViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	//generic build opts
	util.BindViper(v, aprefix+"datadir")
	util.BindViper(v, aprefix+"certsdir")
	//config service
	util.BindViper(v, aprefix+"service.dirs")
	util.BindViper(v, aprefix+"service.files")
}

// FromViper fill values from viper
func (cfg *XListCfg) FromViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	cfg.ServiceDirs = v.GetStringSlice(aprefix + "service.dirs")
	cfg.ServiceFiles = v.GetStringSlice(aprefix + "service.files")
	cfg.DataDir = v.GetString(aprefix + "datadir")
	cfg.CertsDir = v.GetString(aprefix + "certsdir")
}

// Empty returns true if configuration is empty
func (cfg XListCfg) Empty() bool {
	if len(cfg.ServiceDirs) > 0 {
		return false
	}
	if len(cfg.ServiceFiles) > 0 {
		return false
	}
	if cfg.DataDir != "" {
		return false
	}
	if cfg.CertsDir != "" {
		return false
	}
	return true
}

// Validate checks that configuration is ok
func (cfg XListCfg) Validate() error {
	if len(cfg.ServiceFiles) == 0 && len(cfg.ServiceDirs) == 0 {
		return errors.New("service config required")
	}
	for _, file := range cfg.ServiceFiles {
		if !strings.HasSuffix(file, ".json") {
			return fmt.Errorf("config file '%s' without .json extension", file)
		}
		if !util.FileExists(file) {
			return fmt.Errorf("config file '%v' doesn't exists", file)
		}
	}
	for _, dir := range cfg.ServiceDirs {
		if !util.DirExists(dir) {
			return fmt.Errorf("config dir '%v' doesn't exists", dir)
		}
	}
	if cfg.DataDir != "" {
		if !util.DirExists(cfg.DataDir) {
			return fmt.Errorf("sources dir '%v' doesn't exists", cfg.DataDir)
		}
	}
	if cfg.CertsDir != "" {
		if !util.DirExists(cfg.CertsDir) {
			return fmt.Errorf("certificates dir '%v' doesn't exists", cfg.CertsDir)
		}
	}
	return nil
}

// Dump configuration
func (cfg XListCfg) Dump() string {
	return fmt.Sprintf("%+v", cfg)
}
