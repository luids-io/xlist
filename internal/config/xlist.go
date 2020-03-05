// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package config

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/luids-io/common/util"
)

// XListCfg stores lists config paths and builder prefs
type XListCfg struct {
	//source configuration
	ConfigDirs  []string
	ConfigFiles []string
	//build opts
	SourcesDir string
	CertsDir   string
	DNSxL      DNSxLCfg
}

// SetPFlags setups posix flags for commandline configuration
func (cfg *XListCfg) SetPFlags(short bool, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	if short {
		pflag.StringSliceVarP(&cfg.ConfigDirs, aprefix+"dirs", "D", cfg.ConfigDirs, "Definition config dirs.")
		pflag.StringSliceVarP(&cfg.ConfigFiles, aprefix+"files", "d", cfg.ConfigFiles, "Definition config files.")
	} else {
		pflag.StringSliceVar(&cfg.ConfigDirs, aprefix+"dirs", cfg.ConfigDirs, "Definition config dirs.")
		pflag.StringSliceVar(&cfg.ConfigFiles, aprefix+"files", cfg.ConfigFiles, "Definition config files.")
	}
	pflag.StringVar(&cfg.SourcesDir, aprefix+"sourcesdir", cfg.SourcesDir, "Path to sources files.")
	pflag.StringVar(&cfg.CertsDir, aprefix+"certsdir", cfg.CertsDir, "Path to certificate files.")
	//DNSxL flags
	pflag.IntVar(&cfg.DNSxL.TimeoutMSecs, aprefix+"dnsxl.timeout", cfg.DNSxL.TimeoutMSecs, "DNS timeout in milliseconds.")
	pflag.StringSliceVar(&cfg.DNSxL.Resolvers, aprefix+"dnsxl.resolvers", cfg.DNSxL.Resolvers, "DNS IP resolvers.")
	pflag.BoolVar(&cfg.DNSxL.UseResolvConf, aprefix+"dnsxl.resolvconf", cfg.DNSxL.UseResolvConf, "DNS resolvers from /etc/resolv.conf.")
}

// BindViper setups posix flags for commandline configuration and bind to viper
func (cfg *XListCfg) BindViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	//config
	util.BindViper(v, aprefix+"dirs")
	util.BindViper(v, aprefix+"files")
	//build opts
	util.BindViper(v, aprefix+"sourcesdir")
	util.BindViper(v, aprefix+"certsdir")
	//dnsxl
	util.BindViper(v, aprefix+"dnsxl.timeout")
	util.BindViper(v, aprefix+"dnsxl.resolvers")
	util.BindViper(v, aprefix+"dnsxl.resolvconf")
}

// FromViper fill values from viper
func (cfg *XListCfg) FromViper(v *viper.Viper, prefix string) {
	aprefix := ""
	if prefix != "" {
		aprefix = prefix + "."
	}
	cfg.ConfigDirs = v.GetStringSlice(aprefix + "dirs")
	cfg.ConfigFiles = v.GetStringSlice(aprefix + "files")
	cfg.SourcesDir = v.GetString(aprefix + "sourcesdir")
	cfg.CertsDir = v.GetString(aprefix + "certsdir")
	//dnsxl
	cfg.DNSxL.TimeoutMSecs = viper.GetInt(aprefix + "dnsxl.timeout")
	cfg.DNSxL.Resolvers = viper.GetStringSlice(aprefix + "dnsxl.resolvers")
	cfg.DNSxL.UseResolvConf = viper.GetBool(aprefix + "dnsxl.resolvconf")
}

// Empty returns true if configuration is empty
func (cfg XListCfg) Empty() bool {
	if len(cfg.ConfigDirs) > 0 {
		return false
	}
	if len(cfg.ConfigFiles) > 0 {
		return false
	}
	if cfg.SourcesDir != "" {
		return false
	}
	if cfg.CertsDir != "" {
		return false
	}
	if !cfg.DNSxL.Empty() {
		return false
	}
	return true
}

// Validate checks that configuration is ok
func (cfg XListCfg) Validate() error {
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
	if cfg.SourcesDir != "" {
		if !util.DirExists(cfg.SourcesDir) {
			return fmt.Errorf("sources dir '%v' doesn't exists", cfg.SourcesDir)
		}
	}
	if cfg.CertsDir != "" {
		if !util.DirExists(cfg.CertsDir) {
			return fmt.Errorf("certificates dir '%v' doesn't exists", cfg.CertsDir)
		}
	}
	if !cfg.DNSxL.Empty() {
		err := cfg.DNSxL.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

// Dump configuration
func (cfg XListCfg) Dump() string {
	return fmt.Sprintf("%+v", cfg)
}

// DNSxLCfg stores dnsxl module preferences
type DNSxLCfg struct {
	TimeoutMSecs  int
	Resolvers     []string
	UseResolvConf bool
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
