// Copyright 2020 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package config

import (
	"fmt"

	"github.com/luids-io/common/util"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// SBLookupCfg stores sblookupxl module preferences
type SBLookupCfg struct {
	APIKey    string
	ServerURL string
}

// SetPFlags setups posix flags for commandline configuration
func (cfg *SBLookupCfg) SetPFlags(short bool, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	pflag.StringVar(&cfg.APIKey, aprefix+"apikey", cfg.APIKey, "Safe browsing api key.")
	pflag.StringVar(&cfg.ServerURL, aprefix+"serverurl", cfg.ServerURL, "Safe browsing server url.")
}

// BindViper setups posix flags for commandline configuration and bind to viper
func (cfg *SBLookupCfg) BindViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	util.BindViper(v, aprefix+"apikey")
	util.BindViper(v, aprefix+"serverurl")
}

// FromViper fill values from viper
func (cfg *SBLookupCfg) FromViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	cfg.APIKey = v.GetString(aprefix + "apikey")
	cfg.ServerURL = v.GetString(aprefix + "serverurl")
}

// Empty returns true if configuration is empty
func (cfg SBLookupCfg) Empty() bool {
	if cfg.APIKey != "" || cfg.ServerURL != "" {
		return false
	}
	return true
}

// Validate checks that configuration is ok
func (cfg SBLookupCfg) Validate() error {
	return nil
}

// Dump configuration
func (cfg SBLookupCfg) Dump() string {
	return fmt.Sprintf("%+v", cfg)
}
