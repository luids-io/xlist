// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/luids-io/common/util"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// XLGetCfg stores xlget preferences
type XLGetCfg struct {
	ConfigDirs  []string
	ConfigFiles []string
	OutputDir   string
	CacheDir    string
}

// SetPFlags setups posix flags for commandline configuration
func (cfg *XLGetCfg) SetPFlags(short bool, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	if short {
		pflag.StringSliceVarP(&cfg.ConfigDirs, aprefix+"dirs", "S", cfg.ConfigDirs, "Source dirs.")
		pflag.StringSliceVarP(&cfg.ConfigFiles, aprefix+"files", "s", cfg.ConfigFiles, "Source files.")
	} else {
		pflag.StringSliceVar(&cfg.ConfigDirs, aprefix+"dirs", cfg.ConfigDirs, "Source dirs.")
		pflag.StringSliceVar(&cfg.ConfigFiles, aprefix+"files", cfg.ConfigFiles, "Source files.")
	}
	pflag.StringVar(&cfg.OutputDir, aprefix+"output", cfg.OutputDir, "Output dir.")
	pflag.StringVar(&cfg.CacheDir, aprefix+"cache", cfg.CacheDir, "Cache dir.")
}

// BindViper setups posix flags for commandline configuration and bind to viper
func (cfg *XLGetCfg) BindViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	util.BindViper(v, aprefix+"dirs")
	util.BindViper(v, aprefix+"files")
	util.BindViper(v, aprefix+"output")
	util.BindViper(v, aprefix+"cache")
}

// FromViper fill values from viper
func (cfg *XLGetCfg) FromViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	cfg.ConfigDirs = v.GetStringSlice(aprefix + "dirs")
	cfg.ConfigFiles = v.GetStringSlice(aprefix + "files")
	cfg.OutputDir = v.GetString(aprefix + "output")
	cfg.CacheDir = v.GetString(aprefix + "cache")
}

// Empty returns true if configuration is empty
func (cfg XLGetCfg) Empty() bool {
	if len(cfg.ConfigDirs) > 0 {
		return false
	}
	if len(cfg.ConfigFiles) > 0 {
		return false
	}
	return true
}

// Validate checks that configuration is ok
func (cfg XLGetCfg) Validate() error {
	if len(cfg.ConfigFiles) == 0 && len(cfg.ConfigDirs) == 0 {
		return errors.New("config required")
	}
	for _, file := range cfg.ConfigFiles {
		if !strings.HasSuffix(file, ".json") {
			return fmt.Errorf("config file '%s' without .json extension", file)
		}
		if !util.FileExists(file) {
			return fmt.Errorf("config file '%v' doesn't exists", file)
		}
	}
	for _, dir := range cfg.ConfigDirs {
		if !util.DirExists(dir) {
			return fmt.Errorf("config dir '%v' doesn't exists", dir)
		}
	}
	if cfg.OutputDir != "" {
		if !util.DirExists(cfg.OutputDir) {
			return errors.New("output dir doesn't exists")
		}
	}
	if cfg.CacheDir != "" {
		if !util.DirExists(cfg.CacheDir) {
			return errors.New("cache dir doesn't exists")
		}
	}
	return nil
}

// Dump configuration
func (cfg XLGetCfg) Dump() string {
	return fmt.Sprintf("%+v", cfg)
}
