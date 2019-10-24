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
		var err error
		switch {
		case strings.HasPrefix(source.URI, "http://"):
			err = c.doDownloadHTTP(r, source.URI, source.Filename)
		case strings.HasPrefix(source.URI, "https://"):
			err = c.doDownloadHTTP(r, source.URI, source.Filename)
		case strings.HasPrefix(source.URI, "file://"):
			err = c.doDownloadFile(r, source.URI)
		default:
			err = fmt.Errorf("invalid URI format '%s'", source.URI)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) doDownloadHTTP(r *Response, uri, filename string) error {
	dst := r.tempDir
	if filename != "" {
		dst = dst + string(os.PathSeparator) + filename
	}
	//do download
	req, err := grab.NewRequest(dst, uri)
	if err != nil {
		return err
	}
	resp := c.httpclient.Do(req)
	select {
	case <-resp.Done:
		if resp.Err() != nil {
			return resp.Err()
		}
		r.downloadFiles = append(r.downloadFiles, resp.Filename)
		return nil
	case <-r.stop:
		resp.Cancel()
		return ErrCanceled
	}
}

func (c *Client) doDownloadFile(r *Response, uri string) error {
	file := strings.TrimPrefix(uri, "file://")
	_, err := os.Stat(file)
	if err != nil {
		return err
	}
	//no copy file, only sets as downloaded
	r.downloadFiles = append(r.downloadFiles, file)
	return nil
}

//ValidURI returns true if is a valid uri for downloader
func ValidURI(uri string) bool {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return false
	}
	switch u.Scheme {
	case "http":
		return true
	case "https":
		return true
	case "file":
		return true
	}
	return false
}
