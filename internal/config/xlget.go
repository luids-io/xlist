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

// XLGetCfg stores xlget preferences
type XLGetCfg struct {
	SourceDirs  []string
	SourceFiles []string
	OutputDir   string
	CacheDir    string
	StatusDir   string
}

// SetPFlags setups posix flags for commandline configuration
func (cfg *XLGetCfg) SetPFlags(short bool, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	if short {
		pflag.StringSliceVarP(&cfg.SourceDirs, aprefix+"source.dirs", "S", cfg.SourceDirs, "Source dirs.")
		pflag.StringSliceVarP(&cfg.SourceFiles, aprefix+"source.files", "s", cfg.SourceFiles, "Source files.")
	} else {
		pflag.StringSliceVar(&cfg.SourceDirs, aprefix+"source.dirs", cfg.SourceDirs, "Source dirs.")
		pflag.StringSliceVar(&cfg.SourceFiles, aprefix+"source.files", cfg.SourceFiles, "Source files.")
	}
	pflag.StringVar(&cfg.OutputDir, aprefix+"outputdir", cfg.OutputDir, "Output dir.")
	pflag.StringVar(&cfg.CacheDir, aprefix+"cachedir", cfg.CacheDir, "Cache dir.")
	pflag.StringVar(&cfg.CacheDir, aprefix+"statusdir", cfg.CacheDir, "Status dir.")
}

// BindViper setups posix flags for commandline configuration and bind to viper
func (cfg *XLGetCfg) BindViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	util.BindViper(v, aprefix+"source.dirs")
	util.BindViper(v, aprefix+"source.files")
	util.BindViper(v, aprefix+"outputdir")
	util.BindViper(v, aprefix+"cachedir")
	util.BindViper(v, aprefix+"statusdir")
}

// FromViper fill values from viper
func (cfg *XLGetCfg) FromViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	cfg.SourceDirs = v.GetStringSlice(aprefix + "source.dirs")
	cfg.SourceFiles = v.GetStringSlice(aprefix + "source.files")
	cfg.OutputDir = v.GetString(aprefix + "outputdir")
	cfg.CacheDir = v.GetString(aprefix + "cachedir")
	cfg.StatusDir = v.GetString(aprefix + "statusdir")
}

// Empty returns true if configuration is empty
func (cfg XLGetCfg) Empty() bool {
	if len(cfg.SourceDirs) > 0 {
		return false
	}
	if len(cfg.SourceFiles) > 0 {
		return false
	}
	return true
}

// Validate checks that configuration is ok
func (cfg XLGetCfg) Validate() error {
	if len(cfg.SourceFiles) == 0 && len(cfg.SourceDirs) == 0 {
		return errors.New("config required")
	}
	for _, file := range cfg.SourceFiles {
		if !strings.HasSuffix(file, ".json") {
			return fmt.Errorf("config file '%s' without .json extension", file)
		}
		if !util.FileExists(file) {
			return fmt.Errorf("config file '%v' doesn't exists", file)
		}
	}
	for _, dir := range cfg.SourceDirs {
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
	if cfg.StatusDir != "" {
		if !util.DirExists(cfg.StatusDir) {
			return errors.New("status dir doesn't exists")
		}
	}
	return nil
}

// Dump configuration
func (cfg XLGetCfg) Dump() string {
	return fmt.Sprintf("%+v", cfg)
}
