// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlistd

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/luids-io/api/xlist"

	"github.com/luids-io/core/apiservice"
	"github.com/luids-io/core/yalogi"
)

// Builder constructs RBL services.
type Builder struct {
	opts   builderOpts
	logger yalogi.Logger

	services apiservice.Discover
	lists    map[string]List

	startup  []func() error
	shutdown []func() error
}

// BuildListFn defines a function that constructs a checker.
type BuildListFn func(builder *Builder, parents []string, def ListDef) (List, error)

// BuildWrapperFn defines a function that constructs a wrapper and returns
// the checker wrapped.
type BuildWrapperFn func(builder *Builder, def WrapperDef, list List) (List, error)

// BuilderOption is used for builder configuration.
type BuilderOption func(*builderOpts)

type builderOpts struct {
	certsDir string
	dataDir  string
	logger   yalogi.Logger
}

var defaultOptions = builderOpts{logger: yalogi.LogNull}

// DataDir option sets source dir.
func DataDir(s string) BuilderOption {
	return func(o *builderOpts) {
		o.dataDir = s
	}
}

// CertsDir option sets certificate dir.
func CertsDir(s string) BuilderOption {
	return func(o *builderOpts) {
		o.certsDir = s
	}
}

// SetLogger option sets a logger for the component.
func SetLogger(l yalogi.Logger) BuilderOption {
	return func(o *builderOpts) {
		o.logger = l
	}
}

// NewBuilder instances a new builder.
func NewBuilder(services apiservice.Discover, opt ...BuilderOption) *Builder {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	return &Builder{
		opts:     opts,
		logger:   opts.logger,
		services: services,
		lists:    make(map[string]List),
		startup:  make([]func() error, 0),
		shutdown: make([]func() error, 0),
	}
}

// List returns the RBL created by builder with the id passed as param.
func (b *Builder) List(id string) (List, bool) {
	bl, ok := b.lists[id]
	return bl, ok
}

// Build creates a RBL using the metadata passed as param.
func (b *Builder) Build(def ListDef) (List, error) {
	return b.BuildChild(make([]string, 0), def)
}

// BuildChild allows to create child list for composed RBL.
// Parameter parents is an array with the parents ID and is used for looping
// detection.
func (b *Builder) BuildChild(parents []string, def ListDef) (List, error) {
	b.logger.Debugf("building '%s' class '%s'", def.ID, def.Class)
	if def.ID == "" {
		return nil, errors.New("id field is required")
	}
	//check if is a reused list
	bl, ok := b.lists[def.ID]
	if ok {
		if def.Class != "" {
			return nil, fmt.Errorf("'%s' already exists", def.ID)
		}
		// if has wrappers
		if def.Wrappers != nil && len(def.Wrappers) > 0 {
			var err error
			for _, w := range def.Wrappers {
				bl, err = b.buildWrapper(w, bl)
				if err != nil {
					return nil, fmt.Errorf("building aliased '%s': %v", def.ID, err)
				}
			}
		}
		return bl, nil
	}
	//check if disabled
	if def.Disabled {
		return nil, fmt.Errorf("'%s' is disabled", def.ID)
	}
	// check if removed, returns a removedList instance
	if def.Removed {
		b.logger.Errorf("'%s' is marked as removed", def.ID)
		bl := &removedList{
			id:        def.ID,
			resources: xlist.ClearResourceDups(def.Resources, true),
		}
		//register new created list
		b.lists[def.ID] = bl
		return bl, nil
	}
	// check if deprecated, prints a warning
	if def.Deprecated {
		b.logger.Warnf("'%s' is marked as deprecated", def.ID)
	}
	//check for recursion
	for _, p := range parents {
		if p == def.ID {
			return nil, fmt.Errorf("loop detection in '%s'", def.ID)
		}
	}
	//get builder for related class and construct new list
	customb, ok := regListBuilder[def.Class]
	if !ok {
		return nil, fmt.Errorf("building '%s': can't find a builder for '%s'", def.ID, def.Class)
	}
	bl, err := customb(b, parents, def) //builds list
	if err != nil {
		return nil, fmt.Errorf("building '%s': %v", def.ID, err)
	}
	// create with wrappers
	if def.Wrappers != nil && len(def.Wrappers) > 0 {
		for _, w := range def.Wrappers {
			bl, err = b.buildWrapper(w, bl)
			if err != nil {
				return nil, fmt.Errorf("building '%s': '%s': %v", def.ID, w.Class, err)
			}
		}
	}
	//register new created list
	b.lists[def.ID] = bl
	return bl, nil
}

func (b *Builder) buildWrapper(def WrapperDef, bl List) (List, error) {
	b.logger.Debugf("building '%s' wrapper '%s'", bl.ID(), def.Class)
	customb, ok := regWrapperBuilder[def.Class] //get a builder for related class
	if !ok {
		return nil, errors.New("can't find a builder")
	}
	blc, err := customb(b, def, bl) //builds wrapper
	if err != nil {
		return nil, err
	}
	return blc, nil
}

// OnStartup registers the functions that will be executed during startup.
func (b *Builder) OnStartup(f func() error) {
	b.startup = append(b.startup, f)
}

// OnShutdown registers the functions that will be executed during shutdown.
func (b *Builder) OnShutdown(f func() error) {
	b.shutdown = append(b.shutdown, f)
}

// Start executes all registered functions.
func (b *Builder) Start() error {
	b.logger.Infof("starting xlist-builder registered services")
	var ret error
	for _, f := range b.startup {
		err := f()
		if err != nil {
			return err
		}
	}
	return ret
}

// Shutdown executes all registered functions.
func (b *Builder) Shutdown() error {
	b.logger.Infof("shutting down xlist-builder registered services")
	var ret error
	for _, f := range b.shutdown {
		err := f()
		if err != nil {
			ret = err
		}
	}
	return ret
}

// APIService returns service by name
func (b Builder) APIService(name string) (apiservice.Service, bool) {
	return b.services.GetService(name)
}

// DataPath returns path for source
func (b Builder) DataPath(source string) string {
	if path.IsAbs(source) {
		return source
	}
	output := source
	if b.opts.dataDir != "" {
		output = b.opts.dataDir + string(os.PathSeparator) + output
	}
	return output
}

// CertPath returns path for certificate
func (b Builder) CertPath(cert string) string {
	if path.IsAbs(cert) {
		return cert
	}
	output := cert
	if b.opts.certsDir != "" {
		output = b.opts.certsDir + string(os.PathSeparator) + output
	}
	return output
}

// Logger returns logger
func (b Builder) Logger() yalogi.Logger {
	return b.logger
}

// RegisterListBuilder registers a list builder for a class name
func RegisterListBuilder(class string, builder BuildListFn) {
	regListBuilder[class] = builder
}

// RegisterWrapperBuilder registers a wrapper builder for a class name
func RegisterWrapperBuilder(class string, builder BuildWrapperFn) {
	regWrapperBuilder[class] = builder
}

// Package level registry builders
var regListBuilder map[string]BuildListFn
var regWrapperBuilder map[string]BuildWrapperFn

func init() {
	regListBuilder = make(map[string]BuildListFn)
	regWrapperBuilder = make(map[string]BuildWrapperFn)
}
