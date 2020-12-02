// Copyright 2020 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"fmt"

	checkapi "github.com/luids-io/api/xlist/grpc/check"
	"github.com/luids-io/core/yalogi"
	"github.com/luids-io/xlist/internal/config"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// XListCheckAPI creates grpc service
func XListCheckAPI(cfg *config.XListCheckAPICfg, finder *xlistd.Builder, logger yalogi.Logger) (*checkapi.Service, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("bad config: %v", err)
	}
	list, ok := finder.List(cfg.RootListID)
	if !ok {
		return nil, fmt.Errorf("list '%s' not found", cfg.RootListID)
	}
	if !cfg.Log {
		logger = yalogi.LogNull
	}
	svc := checkapi.NewService(list, checkapi.SetServiceLogger(logger))
	return svc, nil
}
