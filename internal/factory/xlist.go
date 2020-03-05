// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"fmt"

	"github.com/luids-io/common/util"
	"github.com/luids-io/xlist/internal/config"
	"github.com/luids-io/xlist/pkg/listbuilder"
	"github.com/luisguillenc/yalogi"
)

//Lists creates lists from configuration files
func Lists(cfg *config.XListCfg, builder *listbuilder.Builder, logger yalogi.Logger) error {
	err := cfg.Validate()
	if err != nil {
		return fmt.Errorf("bad config: %v", err)
	}
	dbfiles, err := util.GetFilesDB("json", cfg.ConfigFiles, cfg.ConfigDirs)
	if err != nil {
		return fmt.Errorf("loading dbfiles: %v", err)
	}
	defs, err := loadListDefs(dbfiles)
	if err != nil {
		return fmt.Errorf("loading dbfiles: %v", err)
	}
	for _, def := range defs {
		if def.Disabled {
			continue
		}
		_, err := builder.Build(def)
		if err != nil {
			return fmt.Errorf("creating '%s': %v", def.ID, err)
		}
	}
	return nil
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
