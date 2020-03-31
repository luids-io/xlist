// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package config

import (
	"errors"
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/luids-io/common/util"
)

// XListCheckAPICfg stores check service preferences
type XListCheckAPICfg struct {
	RootListID string
	ExposePing bool
	Disclosure bool
}

// SetPFlags setups posix flags for commandline configuration
func (cfg *XListCheckAPICfg) SetPFlags(short bool, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	pflag.StringVar(&cfg.RootListID, aprefix+"rootid", cfg.RootListID, "Root list ID for check service.")
	pflag.BoolVar(&cfg.ExposePing, aprefix+"exposeping", cfg.ExposePing, "Exposes internal ping in the service.")
	pflag.BoolVar(&cfg.Disclosure, aprefix+"disclosure", cfg.Disclosure, "Disclosure internal errors.")
}

// BindViper setups posix flags for commandline configuration and bind to viper
func (cfg *XListCheckAPICfg) BindViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	util.BindViper(v, aprefix+"rootid")
	util.BindViper(v, aprefix+"exposeping")
	util.BindViper(v, aprefix+"disclosure")
}

// FromViper fill values from viper
func (cfg *XListCheckAPICfg) FromViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	cfg.RootListID = v.GetString(aprefix + "rootid")
	cfg.ExposePing = v.GetBool(aprefix + "exposeping")
	cfg.Disclosure = v.GetBool(aprefix + "disclosure")
}

// Empty returns true if configuration is empty
func (cfg XListCheckAPICfg) Empty() bool {
	if cfg.RootListID != "" {
		return false
	}
	return true
}

// Validate checks that configuration is ok
func (cfg XListCheckAPICfg) Validate() error {
	if cfg.RootListID == "" {
		return errors.New("root list can't be empty")
	}
	return nil
}

// Dump configuration
func (cfg XListCheckAPICfg) Dump() string {
	return fmt.Sprintf("%+v", cfg)
}
