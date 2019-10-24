// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"fmt"

	"github.com/luids-io/common/util"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/internal/config"
	listbuilder "github.com/luids-io/xlist/pkg/builder"
)

// RootXList is a factory for an xlist service
func RootXList(cfg *config.XListCfg, builder *listbuilder.Builder) (xlist.Checker, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("bad config: %v", err)
	}
	dbfiles, err := util.GetFilesDB("json", cfg.ConfigFiles, cfg.ConfigDirs)
	if err != nil {
		return nil, fmt.Errorf("loading dbfiles: %v", err)
	}
	listDefs, err := loadListDefs(dbfiles)
	if err != nil {
		return nil, fmt.Errorf("loading dbfiles: %v", err)
	}
	for _, def := range listDefs {
		if def.Disabled {
			continue
		}
		_, err := builder.Build(def)
		if err != nil {
			return nil, fmt.Errorf("creating list '%s': %v", def.ID, err)
		}
	}
	rootList, ok := builder.List(cfg.RootListID)
	if !ok {
		return nil, fmt.Errorf("couldn't get root list '%s'", cfg.RootListID)
	}
	return rootList, nil
}

func loadListDefs(dbFiles []string) ([]listbuilder.ListDef, error) {
	loadedDB := make([]listbuilder.ListDef, 0)
	for _, file := range dbFiles {
		entries, err := listbuilder.DefsFromFile(file)
		if err != nil {
			return nil, fmt.Errorf("couln't load database: %v", err)
		}
		loadedDB = append(loadedDB, entries...)
	}
	return loadedDB, nil
}
