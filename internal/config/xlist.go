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

// XListCfg stores xlist/xlmulti preferences
type XListCfg struct {
	RootListID  string
	ConfigDirs  []string
	ConfigFiles []string

	ExposePing bool
	Disclosure bool
}

// SetPFlags setups posix flags for commandline configuration
func (cfg *XListCfg) SetPFlags(short bool, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	if short {
		pflag.StringSliceVarP(&cfg.ConfigDirs, aprefix+"dirs", "D", cfg.ConfigDirs, "Definition dirs.")
		pflag.StringSliceVarP(&cfg.ConfigFiles, aprefix+"files", "d", cfg.ConfigFiles, "Definition files.")
	} else {
		pflag.StringSliceVar(&cfg.ConfigDirs, aprefix+"dirs", cfg.ConfigDirs, "Definition dirs.")
		pflag.StringSliceVar(&cfg.ConfigFiles, aprefix+"files", cfg.ConfigFiles, "Definition files.")
	}
	pflag.StringVar(&cfg.RootListID, aprefix+"rootid", cfg.RootListID, "Root list ID of the xlist service.")
	pflag.BoolVar(&cfg.ExposePing, aprefix+"exposeping", cfg.ExposePing, "Exposes internal ping in the service.")
	pflag.BoolVar(&cfg.Disclosure, aprefix+"disclosure", cfg.Disclosure, "Disclosure internal errors.")
}

// BindViper setups posix flags for commandline configuration and bind to viper
func (cfg *XListCfg) BindViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	util.BindViper(v, aprefix+"dirs")
	util.BindViper(v, aprefix+"files")
	util.BindViper(v, aprefix+"rootid")
	util.BindViper(v, aprefix+"exposeping")
	util.BindViper(v, aprefix+"disclosure")
}

// FromViper fill values from viper
func (cfg *XListCfg) FromViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	cfg.ConfigDirs = v.GetStringSlice(aprefix + "dirs")
	cfg.ConfigFiles = v.GetStringSlice(aprefix + "files")
	cfg.RootListID = v.GetString(aprefix + "rootid")
	cfg.ExposePing = v.GetBool(aprefix + "exposeping")
	cfg.Disclosure = v.GetBool(aprefix + "disclosure")
}

// Empty returns true if configuration is empty
func (cfg XListCfg) Empty() bool {
	if len(cfg.ConfigDirs) > 0 {
		return false
	}
	if len(cfg.ConfigFiles) > 0 {
		return false
	}
	if cfg.RootListID != "" {
		return false
	}
	return true
}

// Validate checks that configuration is ok
func (cfg XListCfg) Validate() error {
	if cfg.RootListID == "" {
		return errors.New("root list can't be empty")
	}
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
	return nil
}

// Dump configuration
func (cfg XListCfg) Dump() string {
	return fmt.Sprintf("%+v", cfg)
}
