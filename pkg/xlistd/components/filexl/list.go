// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package filexl

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/yalogi"
	"github.com/luids-io/xlist/pkg/xlistd/components/memxl"
)

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Logger:     yalogi.LogNull,
		ReloadTime: 30 * time.Second,
	}
}

// Config options
type Config struct {
	Logger          yalogi.Logger
	ForceValidation bool
	Reason          string
	UnsafeReload    bool
	Autoreload      bool
	ReloadTime      time.Duration
}

type options struct {
	forceValidation bool
	reason          string
	unsafeReload    bool
	autoreload      bool
	reloadTime      time.Duration
}

// List loads list from a file using internally a memxl.List
type List struct {
	id        string
	opts      options
	logger    yalogi.Logger
	filename  string
	resources []xlist.Resource
	// mtime and size are only read and modified by a single goroutine
	mtime time.Time
	size  int64

	list    *memxl.List
	mu      sync.RWMutex
	err     error
	close   chan bool
	started bool
}

// New creates a new List with the filename passed
func New(id, filename string, resources []xlist.Resource, cfg Config) *List {
	l := &List{
		id:       id,
		filename: filename,
		opts: options{
			forceValidation: cfg.ForceValidation,
			reason:          cfg.Reason,
			unsafeReload:    cfg.UnsafeReload,
			autoreload:      cfg.Autoreload,
			reloadTime:      cfg.ReloadTime,
		},
		logger:    cfg.Logger,
		resources: xlist.ClearResourceDups(resources),
		close:     make(chan bool),
	}
	l.list = memxl.New(l.id, l.resources,
		memxl.Config{
			ForceValidation: l.opts.forceValidation,
			Reason:          l.opts.reason,
		})

	if l.logger == nil {
		l.logger = yalogi.LogNull
	}
	return l
}

// ID implements xlistd.List interface
func (l *List) ID() string {
	return l.id
}

// Class implements xlistd.List interface
func (l *List) Class() string {
	return BuildClass
}

// Check implements xlist.Checker interface
func (l *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if !l.started {
		return xlist.Response{}, xlist.ErrUnavailable
	}
	//this mutex is for don't allow reload file while checking in memory
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.list.Check(ctx, name, resource)
}

// Resources implements xlist.Checker interface
func (l *List) Resources() []xlist.Resource {
	resources := make([]xlist.Resource, len(l.resources), len(l.resources))
	copy(resources, l.resources)
	return resources
}

// Ping implements xlist.Checker interface
func (l *List) Ping() error {
	if !l.started {
		return errors.New("list is closed")
	}
	return l.err
}

// ReadOnly implements xlistd.List interface
func (l *List) ReadOnly() bool {
	return true
}

// Open loads file to memory
func (l *List) Open() error {
	l.logger.Debugf("%s: opening source '%s'", l.id, l.filename)
	err := l.loadFile(l.list, false)
	if err != nil {
		return err
	}
	if l.opts.autoreload {
		go l.doReload()
	}
	l.started = true
	return nil
}

// Close release memory
func (l *List) Close() {
	l.logger.Debugf("%s: closing source '%s'", l.id, l.filename)
	if l.started {
		if l.opts.autoreload {
			l.close <- true
		}
		l.started = false
		l.list.Clear()
	}
}

// Reload forces a reload from the file
func (l *List) Reload() error {
	l.logger.Debugf("reloading source '%s'", l.filename)
	if l.opts.unsafeReload {
		l.mu.Lock()
		defer l.mu.Unlock()
		err := l.loadFile(l.list, true)
		if err != nil {
			return err
		}
		return nil
	}
	//safe reload
	list := memxl.New(l.id, l.resources,
		memxl.Config{
			ForceValidation: l.opts.forceValidation,
			Reason:          l.opts.reason,
		})
	err := l.loadFile(list, false)
	if err != nil {
		return err
	}
	//replace hashlist
	old := l.list
	l.mu.Lock()
	l.list = list
	l.mu.Unlock()
	old.Clear()
	return nil
}

func (l *List) doReload() {
	ticker := time.NewTicker(l.opts.reloadTime)
	for {
		select {
		case <-l.close:
			return
		case <-ticker.C:
			l.logger.Debugf("checking source '%s'", l.filename)
			changed, err := l.changed()
			if err == nil && changed {
				l.logger.Infof("source '%s' has changed", l.filename)
				err = l.Reload()
			}
			l.setErr(err)
		}
	}
}

func (l *List) changed() (bool, error) {
	file, err := os.Open(l.filename)
	defer file.Close()
	if err != nil {
		return false, err
	}
	stat, err := file.Stat()
	if err != nil {
		return false, err
	}
	if l.mtime.Equal(stat.ModTime()) && l.size == stat.Size() {
		return false, nil
	}
	return true, nil
}

func (l *List) loadFile(hashlist *memxl.List, clear bool) error {
	file, err := os.Open(l.filename)
	defer file.Close()
	if err != nil {
		return err
	}
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	err = memxl.LoadFromFile(hashlist, l.filename, clear)
	l.mtime = stat.ModTime()
	l.size = stat.Size()
	return err
}

func (l *List) setErr(err error) {
	l.err = err
	if err != nil {
		l.logger.Warnf("%s: %v", l.id, err)
	}
}
