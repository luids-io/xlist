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
	for idx, compressed := range r.downloadFiles {
		c.logger.Debugf("uncompressing '%s'", compressed)
		if idx >= len(r.request.Sources) {
			return fmt.Errorf("unbound index %v", idx)
		}
		format := r.request.Sources[idx].Compression
		u, err := getUncompressor(format)
		if err != nil {
			return err
		}
		err = u.uncompress(r, compressed)
		if err != nil {
			return err
		}
	}
	return nil
}

type uncompressor interface {
	uncompress(r *Response, file string) error
}

func getUncompressor(c Compression) (uncompressor, error) {
	switch c {
	case None:
		return noneUncompressor{}, nil
	case Gzip:
		return gzipUncompressor{}, nil
	case Zip:
		return zipUncompressor{}, nil
	}
	return nil, fmt.Errorf("invalid Compression format '%v'", c)
}

type noneUncompressor struct{}

func (u noneUncompressor) uncompress(r *Response, file string) error {
	_, err := os.Stat(file)
	if err != nil {
		return err
	}
	//no uncompress file, only sets as source
	r.sourceFiles = append(r.sourceFiles, file)
	return nil
}

type gzipUncompressor struct{}

func (g gzipUncompressor) uncompress(r *Response, file string) error {
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

type zipUncompressor struct{}

func (z zipUncompressor) uncompress(r *Response, file string) error {
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
