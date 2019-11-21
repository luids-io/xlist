// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package memxl

import (
	"errors"
	"fmt"
	"os"

	"github.com/luids-io/core/option"
	"github.com/luids-io/core/xlist"
	listbuilder "github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines default class for component builder
const BuildClass = "mem"

// Builder returns a list builder function
func Builder(opt ...Option) listbuilder.BuildCheckerFn {
	return func(builder *listbuilder.Builder, parents []string, list listbuilder.ListDef) (xlist.Checker, error) {
		source := ""
		if list.Source != "" {
			source = builder.SourcePath(list.Source)
			if !fileExists(source) {
				return nil, fmt.Errorf("file '%s' doesn't exists", source)
			}
		}
		var data []Data
		bopt := make([]Option, 0)
		bopt = append(bopt, opt...)
		if list.Opts != nil {
			var err error
			bopt, err = parseOptions(bopt, list.Opts)
			if err != nil {
				return nil, err
			}
			data, err = getData(list.Opts)
			if err != nil {
				return nil, err
			}
		}
		bl := New(list.Resources, bopt...)

		//register startup
		builder.OnStartup(func() error {
			builder.Logger().Debugf("starting '%s'", list.ID)
			if source != "" {
				err := LoadFromFile(bl, source, false)
				if err != nil {
					return err
				}
			}
			if data != nil {
				err := LoadFromData(bl, data, false)
				if err != nil {
					return err
				}
			}
			return nil
		})

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

func parseOptions(bopt []Option, opts map[string]interface{}) ([]Option, error) {
	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, Reason(reason))
	}

	return bopt, nil
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
			format, err := xlist.ToFormat(f)
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
	listbuilder.RegisterCheckerBuilder(BuildClass, Builder())
}
