// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package factory

import (
	"fmt"

	"github.com/luids-io/common/util"
	"github.com/luids-io/core/yalogi"
	"github.com/luids-io/xlist/internal/config"
	"github.com/luids-io/xlist/pkg/xlget"
)

// XLGet is a factory for an xlget manager
func XLGet(cfg *config.XLGetCfg, logger yalogi.Logger) (*xlget.Manager, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("bad config: %v", err)
	}
	dbfiles, err := util.GetFilesDB("json", cfg.SourceFiles, cfg.SourceDirs)
	if err != nil {
		return nil, fmt.Errorf("loading dbfiles: %v", err)
	}
	entries, err := loadEntries(dbfiles)
	if err != nil {
		return nil, fmt.Errorf("loading dbfiles: %v", err)
	}
	manager, err := xlget.NewManager(cfg.OutputDir, cfg.CacheDir, xlget.SetLogger(logger))
	if err != nil {
		return nil, fmt.Errorf("creating manager: %v", err)
	}
	err = manager.Add(entries)
	if err != nil {
		return nil, fmt.Errorf("adding entries: %v", err)
	}
	return manager, err
}

func loadEntries(dbFiles []string) ([]xlget.Entry, error) {
	loadedDB := make([]xlget.Entry, 0)
	for _, file := range dbFiles {
		entries, err := xlget.DefsFromFile(file)
		if err != nil {
			return nil, fmt.Errorf("couln't load database: %v", err)
		}
		loadedDB = append(loadedDB, entries...)
	}
	return loadedDB, nil
}
