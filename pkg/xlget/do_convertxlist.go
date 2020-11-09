// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/yalogi"
	"github.com/luids-io/xlist/pkg/xlistd"
)

type xlistConv struct {
	resources []xlist.Resource
	limit     int
}

func (c xlistConv) convert(ctx context.Context, in io.Reader, out chan<- item, logger yalogi.Logger) error {
	nline, nitems := 0, 0
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ErrCanceled
		default:
		}

		nline++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) < 3 {
			return fmt.Errorf("line %v: invalid format", nline)
		}
		resource, err := xlist.ToResource(fields[0])
		if err != nil {
			return fmt.Errorf("line %v: not valid resource", nline)
		}
		if !c.checks(resource) {
			continue
		}
		format, err := xlistd.ToFormat(fields[1])
		if err != nil {
			return fmt.Errorf("line %v: not valid format", nline)
		}
		out <- item{res: resource, format: format, name: fields[2]}
		nitems++
		if c.limited(nitems) {
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanning input: %v", err)
	}
	return nil
}

func (c xlistConv) checks(r xlist.Resource) bool {
	return r.InArray(c.resources)
}

func (c xlistConv) limited(nitems int) bool {
	return c.limit > 0 && nitems > c.limit
}
