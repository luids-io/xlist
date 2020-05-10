// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/cavaliercoder/grab"

	"github.com/luids-io/core/yalogi"
)

// Client downloads requests
type Client struct {
	outputDir, cacheDir string
	logger              yalogi.Logger
	httpclient          *grab.Client
}

//NewClient returns a new Client
func NewClient(outputDir, cacheDir string, logger yalogi.Logger) *Client {
	c := &Client{
		outputDir:  outputDir,
		cacheDir:   cacheDir,
		logger:     logger,
		httpclient: grab.NewClient(),
	}
	return c
}

//Do request a download
func (c *Client) Do(request Request) (*Response, error) {
	c.logger.Infof("getting '%s'", request.ID)
	//validate request
	err := request.validate()
	if err != nil {
		return nil, err
	}
	//setup output
	if request.Output == "" {
		request.Output = fmt.Sprintf("%s.xlist", request.ID)
	}
	//setup outputdir
	if c.outputDir != "" && !path.IsAbs(request.Output) {
		request.Output = c.outputDir + string(os.PathSeparator) + request.Output
	}
	//create odir
	odir := filepath.Dir(request.Output)
	if odir != "." && !dirExists(odir) {
		c.logger.Debugf("creating dir '%s'", odir)
		err := createDir(odir)
		if err != nil {
			return nil, err
		}
	}
	//setup tdir
	tdir := c.cacheDir + string(os.PathSeparator) + strings.ToLower(request.ID)
	if !dirExists(tdir) {
		c.logger.Debugf("creating dir '%s'", tdir)
		err := createDir(tdir)
		if err != nil {
			return nil, err
		}
	}
	//make response and process
	response := &Response{
		ID:            request.ID,
		Output:        request.Output,
		request:       &request,
		status:        Ready,
		stop:          make(chan bool, 0),
		tempDir:       tdir,
		downloadFiles: make([]string, 0, len(request.Sources)),
		Done:          make(chan struct{}, 0),
		Start:         time.Now(),
		Account:       emptyAccount(),
	}
	go c.doProcess(response)
	return response, nil
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
