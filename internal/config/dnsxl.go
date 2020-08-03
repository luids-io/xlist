// Copyright 2020 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package config

import (
	"fmt"
	"net"

	"github.com/luids-io/common/util"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// DNSxLCfg stores dnsxl module preferences
type DNSxLCfg struct {
	TimeoutMSecs  int
	Resolvers     []string
	UseResolvConf bool
}

// SetPFlags setups posix flags for commandline configuration
func (cfg *DNSxLCfg) SetPFlags(short bool, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	pflag.IntVar(&cfg.TimeoutMSecs, aprefix+"timeout", cfg.TimeoutMSecs, "DNS timeout in milliseconds.")
	pflag.StringSliceVar(&cfg.Resolvers, aprefix+"resolvers", cfg.Resolvers, "DNS IP resolvers.")
	pflag.BoolVar(&cfg.UseResolvConf, aprefix+"resolvconf", cfg.UseResolvConf, "DNS resolvers from /etc/resolv.conf.")
}

// BindViper setups posix flags for commandline configuration and bind to viper
func (cfg *DNSxLCfg) BindViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	util.BindViper(v, aprefix+"timeout")
	util.BindViper(v, aprefix+"resolvers")
	util.BindViper(v, aprefix+"resolvconf")
}

// FromViper fill values from viper
func (cfg *DNSxLCfg) FromViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	cfg.TimeoutMSecs = viper.GetInt(aprefix + "timeout")
	cfg.Resolvers = viper.GetStringSlice(aprefix + "resolvers")
	cfg.UseResolvConf = viper.GetBool(aprefix + "resolvconf")
}

// Empty returns true if configuration is empty
func (cfg DNSxLCfg) Empty() bool {
	if cfg.TimeoutMSecs != 0 {
		return false
	}
	if cfg.Resolvers != nil && len(cfg.Resolvers) > 0 {
		return false
	}
	if cfg.UseResolvConf {
		return false
	}
	return true
}

// Validate checks that configuration is ok
func (cfg DNSxLCfg) Validate() error {
	if cfg.TimeoutMSecs < 0 {
		return fmt.Errorf("dns timeout milliseconds is not valid")
	}
	if cfg.UseResolvConf && cfg.Resolvers != nil && len(cfg.Resolvers) > 0 {
		return fmt.Errorf("useresolvconf and resolvers are incompatible")
	}
	for _, s := range cfg.Resolvers {
		ip := net.ParseIP(s)
		if ip == nil {
			return fmt.Errorf("not a valid ip address '%s'", s)
		}
	}
	return nil
}

// Dump configuration
func (cfg DNSxLCfg) Dump() string {
	return fmt.Sprintf("%+v", cfg)
}
