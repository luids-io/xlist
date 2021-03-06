// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package memxl

import (
	"errors"
	"fmt"
	"os"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/option"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// Builder returns a builder function.
func Builder(defaultCfg Config) xlistd.BuildListFn {
	return func(b *xlistd.Builder, parents []string, def xlistd.ListDef) (xlistd.List, error) {
		cfg := defaultCfg
		source := ""
		if def.Source != "" {
			source = b.DataPath(def.Source)
			if !fileExists(source) {
				return nil, fmt.Errorf("file '%s' doesn't exists", source)
			}
		}
		var data []Data
		if def.Opts != nil {
			var err error
			cfg, err = parseOptions(cfg, def.Opts)
			if err != nil {
				return nil, err
			}
			data, err = getData(def.Opts)
			if err != nil {
				return nil, err
			}
		}
		bl := New(def.ID, def.Resources, cfg)
		if len(data) > 0 {
			err := LoadFromData(bl, data, false)
			if err != nil {
				return nil, err
			}
		}
		return bl, nil
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func parseOptions(src Config, opts map[string]interface{}) (Config, error) {
	dst := src
	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Reason = reason
	}
	return dst, nil
}

func getData(opts map[string]interface{}) ([]Data, error) {
	data := make([]Data, 0)
	value, ok, err := option.SliceHashString(opts, "data")
	if err != nil {
		return data, err
	}
	if ok {
		for _, item := range value {
			r, ok := item["resource"]
			if !ok {
				return data, errors.New("invalid 'data': required 'resource'")
			}
			resource, err := xlist.ToResource(r)
			if err != nil {
				return data, fmt.Errorf("invalid 'data': invalid 'resource': %v", err)
			}
			f, ok := item["format"]
			if !ok {
				return data, errors.New("invalid 'data': required 'format'")
			}
			format, err := xlistd.ToFormat(f)
			if err != nil {
				return data, fmt.Errorf("invalid 'data': invalid 'format': %v", err)
			}
			v, ok := item["value"]
			if !ok {
				return data, errors.New("invalid 'data': required 'value'")
			}
			data = append(data,
				Data{
					Resource: resource,
					Format:   format,
					Value:    v,
				})
		}
	}
	return data, nil
}

func init() {
	xlistd.RegisterListBuilder(ComponentClass, Builder(Config{}))
}
