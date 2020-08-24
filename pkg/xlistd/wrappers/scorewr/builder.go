// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package scorewr

import (
	"fmt"
	"regexp"

	"github.com/luids-io/core/option"
	"github.com/luids-io/xlist/pkg/xlistd"
	"github.com/luids-io/xlist/pkg/xlistd/builder"
)


// Builder returns a builder for the component
func Builder(defaultCfg Config) builder.BuildWrapperFn {
	return func(b *builder.Builder, def builder.WrapperDef, list xlistd.List) (xlistd.List, error) {
		cfg := defaultCfg
		score := 0
		if def.Opts != nil {
			v, ok, err := option.Int(def.Opts, "value")
			if err != nil {
				return nil, err
			}
			if ok {
				score = v
			}
			cfg, err = parseOptions(cfg, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		return New(list, score, cfg), nil
	}
}

func parseOptions(src Config, opts map[string]interface{}) (Config, error) {
	dst := src
	matches, ok, err := option.SliceHash(opts, "matches")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Scores = make([]ScoreExpr, 0, len(matches))
		for _, match := range matches {
			expr, ok, err := option.String(match, "expr")
			if err != nil || !ok {
				return dst, err
			}
			value, ok, err := option.Int(match, "value")
			if err != nil || !ok {
				return dst, err
			}
			// compile regexpr
			re, err := regexp.Compile(expr)
			if err != nil {
				return dst, fmt.Errorf("invalid 'matches': invalid 'expr': %s %v", expr, err)
			}
			dst.Scores = append(dst.Scores, ScoreExpr{RegExp: re, Score: value})
		}
	}
	return dst, nil
}

func init() {
	builder.RegisterWrapperBuilder(WrapperClass, Builder(Config{}))
}
