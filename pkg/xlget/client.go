// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/luids-io/core/yalogi"
)

// Common errors
var (
	ErrCanceled = errors.New("download canceled")
)

// Config defines configuration.
type Config struct {
	OutputDir string
	CacheDir  string
	Logger    yalogi.Logger
}

// setDefaults configures Config to have default parameters.
// It reports whether the current configuration is valid.
func (c *Config) setDefaults() bool {
	if c.OutputDir == "" {
		c.OutputDir = defaultOutputDir()
	}
	if c.CacheDir == "" {
		c.CacheDir = defaultCacheDir()
	}
	if c.Logger == nil {
		c.Logger = yalogi.LogNull
	}
	return true
}

// Client downloads requests
type Client struct {
	outputDir string
	cacheDir  string
	logger    yalogi.Logger
}

//NewClient returns a new Client
func NewClient(cfg Config) *Client {
	cfg.setDefaults()
	return &Client{
		outputDir: cfg.OutputDir,
		cacheDir:  cfg.CacheDir,
		logger:    cfg.Logger,
	}
}

//Do request a download
func (c *Client) Do(e Entry) (*Response, error) {
	c.logger.Infof("getting '%s'", e.ID)
	// validate request
	err := ValidateEntry(e)
	if err != nil {
		return nil, err
	}
	// make response
	r := &Response{
		ID:      e.ID,
		request: e.Copy(),
	}
	// setup output
	err = c.setupOutput(r)
	if err != nil {
		return nil, err
	}
	// setup cache
	err = c.setupCache(r)
	if err != nil {
		return nil, err
	}
	// setup response
	r.logger = c.logger
	r.stop = make(chan bool, 0)
	r.downloadFiles = make([]string, 0, len(e.Sources))
	r.Done = make(chan struct{}, 0)
	r.Start = time.Now()
	r.Account = emptyAccount()

	go c.doProcess(r)
	return r, nil
}

func (c *Client) setupOutput(r *Response) error {
	var err error
	//setup output
	if r.request.Output != "" {
		if c.outputDir != "" && !path.IsAbs(r.request.Output) {
			r.Output = c.outputDir + string(os.PathSeparator) + r.request.Output
		}
	} else {
		ouputFile := fmt.Sprintf("%s.xlist", r.ID)
		if c.outputDir != "" {
			r.Output = c.outputDir + string(os.PathSeparator) + ouputFile
		}
	}
	//create setup dir
	r.outputDir = filepath.Dir(r.Output)
	if r.outputDir != "." && !dirExists(r.outputDir) {
		c.logger.Debugf("creating dir '%s'", r.outputDir)
		err = createDir(r.outputDir)
	}
	return err
}

func (c *Client) setupCache(r *Response) error {
	var err error
	if !dirExists(c.cacheDir) {
		c.logger.Debugf("creating dir '%s'", r.tempDir)
		err = createDir(c.cacheDir)
		if err != nil {
			return err
		}
	}
	//setup cache dir
	r.tempDir = c.cacheDir + string(os.PathSeparator) + strings.ToLower(r.ID)
	if !dirExists(r.tempDir) {
		c.logger.Debugf("creating dir '%s'", r.tempDir)
		err = createDir(r.tempDir)
	}
	return err
}

// main process
func (c *Client) doProcess(r *Response) {
	//Download files
	r.status = Downloading
	err := c.doDownload(r)
	if err != nil {
		r.err = ErrCanceled
		if err != ErrCanceled {
			r.err = fmt.Errorf("downloading: %v", err)
		}
		goto FINISH
	}
	//Uncompress files
	r.status = Uncompressing
	err = c.doUncompress(r)
	if err != nil {
		r.err = ErrCanceled
		if err != ErrCanceled {
			r.err = fmt.Errorf("uncompressing: %v", err)
		}
		goto FINISH
	}
	//Convert files
	r.status = Converting
	err = c.doConvert(r)
	if err != nil {
		r.err = ErrCanceled
		if err != ErrCanceled {
			r.err = fmt.Errorf("converting: %v", err)
		}
		goto FINISH
	}
	//Deploy list
	r.status = Deploying
	r.Updated, err = c.doDeploy(r)
	if err != nil {
		r.err = ErrCanceled
		if err != ErrCanceled {
			r.err = fmt.Errorf("deploying: %v", err)
		}
		goto FINISH
	}
	//Clean downloaded files
	if !r.request.NoClean {
		r.status = Cleaning
		err = c.doClean(r)
		if err != nil {
			r.err = ErrCanceled
			if err != ErrCanceled {
				r.err = fmt.Errorf("cleaning: %v", err)
			}
		}
	}
FINISH:
	r.End = time.Now()
	r.status = Finished
	close(r.stop)
	close(r.Done)
}

func defaultOutputDir() string {
	pwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return pwd
}

func defaultCacheDir() string {
	pwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return pwd + string(os.PathSeparator) + ".cache"
}
