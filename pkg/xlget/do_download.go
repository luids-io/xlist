// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/cavaliercoder/grab"
)

func (c *Client) doDownload(r *Response) error {
	for _, source := range r.request.Sources {
		c.logger.Debugf("downloading '%s'", source.URI)
		//gets downloader
		u, _ := url.ParseRequestURI(source.URI)
		d, err := getDownloader(u.Scheme)
		if err != nil {
			return err
		}
		// do download
		err = d.download(r, source)
		if err != nil {
			return err
		}
	}
	return nil
}

type downloader interface {
	download(r *Response, source Source) error
}

func getDownloader(scheme string) (downloader, error) {
	switch scheme {
	case "http":
		return httpFiledown{httpclient: httpclient}, nil
	case "https":
		return httpFiledown{httpclient: httpclient}, nil
	case "file":
		return localFiledown{}, nil
	}
	return nil, fmt.Errorf("no downloader available for '%s' scheme", scheme)
}

type localFiledown struct{}

func (d localFiledown) download(r *Response, s Source) error {
	file := strings.TrimPrefix(s.URI, "file://")
	_, err := os.Stat(file)
	if err != nil {
		return err
	}
	r.logger.Debugf("set '%s' as downloaded file", file)
	//no copy file, only sets as downloaded
	r.downloadFiles = append(r.downloadFiles, file)
	return nil
}

type httpFiledown struct {
	httpclient *grab.Client
}

func (h httpFiledown) download(r *Response, s Source) error {
	uri, filename := s.URI, s.Filename
	dst := r.tempDir
	if filename != "" {
		dst = dst + string(os.PathSeparator) + filename
	}
	//do download
	r.logger.Debugf("downloading '%s' with grab", uri)
	req, err := grab.NewRequest(dst, uri)
	if err != nil {
		return err
	}
	resp := h.httpclient.Do(req)
	select {
	case <-resp.Done:
		if resp.Err() != nil {
			return resp.Err()
		}
		r.logger.Debugf("set '%s' as downloaded file", resp.Filename)
		r.downloadFiles = append(r.downloadFiles, resp.Filename)
		return nil
	case <-r.stop:
		resp.Cancel()
		return ErrCanceled
	}
}

//singleton http client
var httpclient *grab.Client

// SetHTTPClient sets http client
func SetHTTPClient(h *grab.Client) {
	httpclient = h
}

func init() {
	httpclient = grab.NewClient()
}
