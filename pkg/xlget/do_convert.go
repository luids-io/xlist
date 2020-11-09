// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/yalogi"
	"github.com/luids-io/xlist/pkg/xlistd"
	"golang.org/x/net/publicsuffix"
)

func (c *Client) doConvert(r *Response) error {
	//create output tempfile
	r.converted = r.Output + ".tmp"
	outfile, err := os.Create(r.converted)
	if err != nil {
		return fmt.Errorf("can't create converted file: %v", err)
	}
	//convert source files in a goroutine
	exitErr := make(chan error)
	dataCh := make(chan item, 1024)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go convertFiles(ctx, r, dataCh, exitErr)
	//process data
PROCESSDATA:
	for {
		select {
		case <-r.stop:
			cancel()
			break PROCESSDATA
		case data, ok := <-dataCh:
			if !ok {
				break PROCESSDATA
			}
			transformData(r, data, outfile)
		}
	}
	outfile.Close()
	err = <-exitErr
	close(exitErr)
	return err
}

func convertFiles(ctx context.Context, r *Response, dataCh chan<- item, exitErr chan<- error) {
	var err error
	for idx, file := range r.sourceFiles {
		// open file
		r.logger.Debugf("converting file '%s'", file)
		if idx >= len(r.request.Sources) {
			err = fmt.Errorf("unbound index %v", idx)
			break
		}
		var infile *os.File
		infile, err = os.Open(file)
		if err != nil {
			err = fmt.Errorf("can't open: %v", err)
			break
		}
		// get converter
		var c converter
		c, err = getConverter(r.request.Sources[idx])
		if err != nil {
			infile.Close()
			break
		}
		err = c.convert(ctx, infile, dataCh, r.logger)
		if err != nil {
			err = fmt.Errorf("file '%s': %v", file, err)
			infile.Close()
			break
		}
		infile.Close()
	}
	close(dataCh)
	exitErr <- err
}

func transformData(r *Response, data item, outfile io.Writer) {
	if r.request.Transforms != nil {
		if data.res == xlist.Domain && r.request.Transforms.TLDPlusOne {
			tldPlusOne, err := publicsuffix.EffectiveTLDPlusOne(data.name)
			if err == nil && data.name == tldPlusOne {
				data.format = xlistd.Sub
			}
		}
	}
	r.Account[data.res]++
	fmt.Fprintf(outfile, "%s,%s,%s\n", data.res, data.format, data.name)
}

type item struct {
	res    xlist.Resource
	format xlistd.Format
	name   string
}

type converter interface {
	convert(ctx context.Context, in io.Reader, out chan<- item, logger yalogi.Logger) error
}

func getConverter(s Source) (converter, error) {
	switch s.Format {
	case XList:
		return xlistConv{resources: s.Resources, limit: s.Limit}, nil
	case Flat:
		return flatConv{resources: s.Resources, limit: s.Limit}, nil
	case CSV:
		c := csvConv{resources: s.Resources, limit: s.Limit, comma: ',', comment: '#'}
		if s.FormatOpts != nil {
			c.lazyQuotes = s.FormatOpts.LazyQuotes
			c.indexes = s.FormatOpts.Indexes
			c.hasHeader = s.FormatOpts.HasHeader
			if s.FormatOpts.Comma != "" {
				runes := []rune(s.FormatOpts.Comma)
				if len(runes) > 0 {
					c.comma = runes[0]
				}
			}
			if s.FormatOpts.Comment != "" {
				runes := []rune(s.FormatOpts.Comment)
				if len(runes) > 0 {
					c.comment = runes[0]
				}
			}
		}
		return c, nil
	case Hosts:
		return hostsConv{resources: s.Resources, limit: s.Limit}, nil
	}
	return nil, fmt.Errorf("can't locate converter for '%s'", s.Format)
}
