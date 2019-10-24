// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/luisguillenc/yalogi"

	"github.com/luids-io/core/xlist"
)

// Manager processes configuration entries and checks for required updates
type Manager struct {
	outputDir string
	cacheDir  string
	entries   []Entry
	ids       map[string]bool
	c         *Client
	logger    yalogi.Logger
	mu        sync.RWMutex
	running   bool
}

// NewManager creates a new manager
func NewManager(outputDir, cacheDir string, opt ...Option) (*Manager, error) {
	//sets default options
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	if cacheDir == "" {
		cacheDir = defaultCacheDir()
	}
	m := &Manager{
		cacheDir:  cacheDir,
		outputDir: outputDir,
		logger:    opts.logger,
		entries:   make([]Entry, 0),
		ids:       make(map[string]bool),
		c:         NewClient(outputDir, cacheDir, opts.logger),
	}
	return m, m.initDirs()
}

// Option is used for manager configuration
type Option func(*options)

type options struct {
	logger yalogi.Logger
}

var defaultOptions = options{
	logger: yalogi.LogNull,
}

// SetLogger option allows set a custom logger
func SetLogger(l yalogi.Logger) Option {
	return func(o *options) {
		o.logger = l
	}
}

func defaultCacheDir() string {
	pwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return pwd + string(os.PathSeparator) + ".cache"
}

// Add entries to manager
func (m *Manager) Add(entries []Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range entries {
		err := e.Validate()
		if err != nil {
			return err
		}
		_, ok := m.ids[e.ID]
		if ok {
			return fmt.Errorf("duplicated entry id '%s'", e.ID)
		}
		m.entries = append(m.entries, e)
		m.ids[e.ID] = true
	}
	return nil
}

// Clear entries
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = make([]Entry, 0)
	m.ids = make(map[string]bool)
}

// NeedsUpdate returns an slice with ids that must be updated
func (m *Manager) NeedsUpdate() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	needs := make([]string, 0, len(m.entries))
	for _, e := range m.requiresUpdate() {
		needs = append(needs, e.ID)
	}
	return needs
}

// CancelFunc defines a type for cancelation function
type CancelFunc func()

// Update entries registered in backbround, it returns a cancelation function,
// a done channel (it will close when process done) and an error.
func (m *Manager) Update() (CancelFunc, <-chan struct{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.running {
		return nil, nil, errors.New("update already running")
	}
	m.running = true

	requests := make([]Request, 0, len(m.entries))
	for _, e := range m.requiresUpdate() {
		req, err := e.Request()
		if err != nil {
			return nil, nil, err
		}
		requests = append(requests, req)
	}
	// async control
	done := make(chan struct{})
	closeCh := make(chan struct{})
	stop := func() {
		close(closeCh)
		<-done
	}
	go m.updateRequests(requests, closeCh, done)

	return stop, done, nil
}

func (m *Manager) updateRequests(requests []Request, closeCh <-chan struct{}, done chan<- struct{}) {
LOOPREQUESTS:
	for _, req := range requests {
		response, err := m.c.Do(req)
		if err != nil {
			m.logger.Errorf("processing '%s': %v", req.ID, err)
			continue
		}
		select {
		case <-response.Done:
			err = response.Err()
			if err != nil {
				m.logger.Errorf("in response from '%s': %v", req.ID, err)
				continue
			}
			summary := fmt.Sprintf("summary '%s': updated=%v", response.ID, response.Updated)
			for _, r := range xlist.Resources {
				summary = summary + " " + fmt.Sprintf("%v=%v", r, response.Account[r])
			}
			m.logger.Infof("%s\n", summary)
		case <-closeCh:
			response.Cancel()
			response.Wait()
			break LOOPREQUESTS
		}
	}
	m.running = false
	close(done)
}

func (m *Manager) requiresUpdate() []Entry {
	required := make([]Entry, 0, len(m.entries))
	for _, e := range m.entries {
		if e.Disabled {
			continue
		}
		if m.isUpdated(e) {
			continue
		}
		required = append(required, e)
	}
	return required
}

func (m *Manager) getOutputFile(e Entry) string {
	output := e.Output
	if output == "" {
		output = fmt.Sprintf("%s.xlist", e.ID)
	}
	if m.outputDir != "" && !path.IsAbs(output) {
		output = m.outputDir + string(os.PathSeparator) + output
	}
	return output
}

// isUpdated use md5 file time
func (m *Manager) isUpdated(e Entry) bool {
	output := m.getOutputFile(e)
	info, err := os.Stat(output)
	if os.IsNotExist(err) {
		return false
	}
	last := info.ModTime()

	md5file := fmt.Sprintf("%s.md5", output)
	md5info, err := os.Stat(md5file)
	if !os.IsNotExist(err) {
		last = md5info.ModTime()
	}

	now := time.Now()
	modify := now.Sub(last)
	if modify < e.Update.Duration {
		return true
	}
	return false
}

func (m *Manager) initDirs() error {
	if m.outputDir != "" {
		err := createDir(m.outputDir)
		if err != nil {
			return err
		}
	}
	if m.cacheDir != "" {
		err := createDir(m.cacheDir)
		if err != nil {
			return err
		}
	}
	return nil
}
