// Copyright 2020 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"fmt"

	"github.com/luisguillenc/yalogi"

	checkapi "github.com/luids-io/api/xlist/check"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/internal/config"
)

// CheckAPIService creates grpc service
func CheckAPIService(cfg *config.APICheckCfg, finder xlist.ListFinder, logger yalogi.Logger) (*checkapi.Service, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("bad config: %v", err)
	}
	list, ok := finder.FindListByID(cfg.RootListID)
	if !ok {
		return nil, fmt.Errorf("list '%s' not found", cfg.RootListID)
	}
	svc := checkapi.NewService(list,
		checkapi.DisclosureErrors(cfg.Disclosure),
		checkapi.ExposePing(cfg.ExposePing),
	)
	return svc, nil
}
