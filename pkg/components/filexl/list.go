// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package filexl

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/luisguillenc/yalogi"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/components/memxl"
)

type options struct {
	logger          yalogi.Logger
	forceValidation bool
	reason          string
	unsafeReload    bool
	autoreload      bool
	reloadSecs      int
}

var defaultOptions = options{
	logger:     yalogi.LogNull,
	reloadSecs: 30,
}

// Option is used for component configuration
type Option func(*options)

// SetLogger sets a logger for the component
func SetLogger(l yalogi.Logger) Option {
	return func(o *options) {
		o.logger = l
	}
}

// ReloadSeconds number of seconds between check if file has changed
func ReloadSeconds(i int) Option {
	return func(o *options) {
		if i > 0 {
			o.reloadSecs = i
		}
	}
}

// Autoreload uses a goroutine for checking changes in file and reloads it
func Autoreload(b bool) Option {
	return func(o *options) {
		o.autoreload = b
	}
}

// UnsafeReload don't use a temporal memxl for reload
func UnsafeReload(b bool) Option {
	return func(o *options) {
		o.unsafeReload = b
	}
}

// ForceValidation forces components to ignore context and validate requests
func ForceValidation(b bool) Option {
	return func(o *options) {
		o.forceValidation = b
	}
}

// Reason sets a fixed reason for component
func Reason(s string) Option {
	return func(o *options) {
		o.reason = s
	}
}

// List loads list from a file using internally a memxl.List
type List struct {
	opts     options
	logger   yalogi.Logger
	filename string
	// mtime and size are only read and modified by a single goroutine
	mtime time.Time
	size  int64

	hashlist *memxl.List
	mu       sync.RWMutex
	err      error
	close    chan bool
	started  bool
}

// New creates a new List with the filename passed
func New(filename string, resources []xlist.Resource, opt ...Option) *List {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	l := &List{
		opts:     opts,
		logger:   opts.logger,
		filename: filename,
		hashlist: memxl.New(resources,
			memxl.ForceValidation(opts.forceValidation),
			memxl.Reason(opts.reason)),
		close: make(chan bool),
	}
	return l
}

// Check implements xlist.Checker interface
func (l *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if !l.started {
		return xlist.Response{}, xlist.ErrListNotAvailable
	}
	//this mutex is for don't allow reload file while checking in memory
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.hashlist.Check(ctx, name, resource)
}

// Resources implements xlist.Checker interface
func (l *List) Resources() []xlist.Resource {
	return l.hashlist.Resources()
}

// Ping implements xlist.Checker interface
func (l *List) Ping() error {
	if !l.started {
		return xlist.ErrListNotAvailable
	}
	return l.err
}

// Start loads file to memory
func (l *List) Start() error {
	l.logger.Debugf("starting list with source '%s'", l.filename)
	err := l.loadFile(l.hashlist, false)
	if err != nil {
		return err
	}
	if l.opts.autoreload {
		go l.doReload()
	}
	l.started = true
	return nil
}

// Shutdown release memory
func (l *List) Shutdown() {
	l.logger.Debugf("shutting down list with source '%s'", l.filename)
	if l.started {
		if l.opts.autoreload {
			l.close <- true
		}
		l.started = false
		l.hashlist.Clear()
	}
}

// Reload forces a reload from the file
func (l *List) Reload() error {
	l.logger.Debugf("reloading source '%s'", l.filename)
	if l.opts.unsafeReload {
		l.mu.Lock()
		defer l.mu.Unlock()
		err := l.loadFile(l.hashlist, true)
		if err != nil {
			return err
		}
		return nil
	}
	//safe reload
	hashlist := memxl.New(l.Resources(),
		memxl.ForceValidation(l.opts.forceValidation),
		memxl.Reason(l.opts.reason))
	err := l.loadFile(hashlist, false)
	if err != nil {
		return err
	}
	//replace hashlist
	old := l.hashlist
	l.mu.Lock()
	l.hashlist = hashlist
	l.mu.Unlock()
	old.Clear()
	return nil
}

func (l *List) doReload() {
	ticker := time.NewTicker(time.Duration(l.opts.reloadSecs) * time.Second)
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
		l.logger.Warnf("%v", err)
	}
}
