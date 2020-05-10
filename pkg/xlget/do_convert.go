// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/yalogi"
)

// Converter interface defines format converters
type Converter interface {
	SetLogger(log yalogi.Logger)
	Convert(ctx context.Context, in io.Reader, out io.Writer) (map[xlist.Resource]int, error)
}

func (c *Client) doConvert(r *Response) error {
	//create tempfile
	tempfile := r.request.Output + ".tmp"
	outfile, err := os.Create(tempfile)
	if err != nil {
		return fmt.Errorf("can't create converted file: %v", err)
	}
	defer outfile.Close()

	//convert source files
	for idx, file := range r.sourceFiles {
		// open file
		c.logger.Debugf("converting file '%s'", file)
		if idx >= len(r.request.Sources) {
			return fmt.Errorf("unbound index %v", idx)
		}
		infile, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("can't open: %v", err)
		}
		// setup context canceling
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		converter := r.request.Sources[idx].Converter
		if converter == nil {
			converter = &XListConv{}
		}
		converter.SetLogger(c.logger)

		// convert file
		var account map[xlist.Resource]int
		finished := make(chan bool, 0)
		go func() {
			account, err = converter.Convert(ctx, infile, outfile)
			finished <- true
		}()
		select {
		case <-r.stop:
			cancel()
			<-finished
			err = ErrCanceled
		case <-finished:
		}
		close(finished)
		infile.Close()
		if err != nil {
			return fmt.Errorf("file '%s': %v", file, err)
		}
		for key, value := range account {
			r.Account[key] = r.Account[key] + value
		}
	}
	r.converted = tempfile
	return nil
}

func emptyAccount() (account map[xlist.Resource]int) {
	account = make(map[xlist.Resource]int, 0)
	for _, r := range xlist.Resources {
		account[r] = 0
	}
	return
}
