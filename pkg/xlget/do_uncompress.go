// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

func (c *Client) doUncompress(r *Response) error {
	//request.Sources
	for idx, compressed := range r.downloadFiles {
		c.logger.Debugf("uncompressing '%s'", compressed)
		if idx >= len(r.request.Sources) {
			return fmt.Errorf("unbound index %v", idx)
		}
		format := r.request.Sources[idx].Compression

		var err error
		switch format {
		case None:
			err = c.doUncompressNone(r, compressed)
		case Gzip:
			err = c.doUncompressGzip(r, compressed)
		case Zip:
			err = c.doUncompressZip(r, compressed)
		default:
			err = fmt.Errorf("invalid Compression format '%v'", format)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) doUncompressNone(r *Response, file string) error {
	_, err := os.Stat(file)
	if err != nil {
		return err
	}
	//no uncompress file, only sets as source
	r.sourceFiles = append(r.sourceFiles, file)
	return nil
}

func (c *Client) doUncompressGzip(r *Response, file string) error {
	gzipfile, err := os.Open(file)
	if err != nil {
		return err
	}
	dstfile := strings.TrimSuffix(file, ".gz")
	if dstfile == file {
		dstfile = fmt.Sprintf("%s-uncompressed", file)
	}
	reader, err := gzip.NewReader(gzipfile)
	if err != nil {
		return err
	}
	defer reader.Close()
	writer, err := os.Create(dstfile)
	if err != nil {
		return err
	}
	defer writer.Close()
	if _, err = io.Copy(writer, reader); err != nil {
		return err
	}
	//no uncompress file, only sets as source
	r.sourceFiles = append(r.sourceFiles, dstfile)
	return nil
}

func (c *Client) doUncompressZip(r *Response, file string) error {
	dstfile := strings.TrimSuffix(file, ".zip")
	if dstfile == file {
		dstfile = fmt.Sprintf("%s-uncompressed", file)
	}
	zipfile, err := zip.OpenReader(file)
	if err != nil {
		return err
	}
	defer zipfile.Close()
	if len(zipfile.File) != 1 {
		return fmt.Errorf("zip file '%s' contains more than one file", file)
	}
	compfile := zipfile.File[0]

	writer, err := os.Create(dstfile)
	if err != nil {
		return err
	}
	defer writer.Close()
	reader, err := compfile.Open()
	if err != nil {
		return err
	}
	defer reader.Close()
	if _, err = io.Copy(writer, reader); err != nil {
		return err
	}
	//no uncompress file, only sets as source
	r.sourceFiles = append(r.sourceFiles, dstfile)
	return nil
}
