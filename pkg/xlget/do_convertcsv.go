// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/yalogi"
	"github.com/luids-io/xlist/pkg/xlistd"
)

type csvConv struct {
	indexes    []int
	resources  []xlist.Resource
	comma      rune
	comment    rune
	lazyQuotes bool
	hasHeader  bool
	limit      int
}

func (p csvConv) convert(ctx context.Context, in io.Reader, out chan<- item, logger yalogi.Logger) error {
	nline, nitems := 0, 0
	reader := csv.NewReader(bufio.NewReader(in))
	reader.Comma = p.comma
	reader.Comment = p.comment
	reader.LazyQuotes = p.lazyQuotes

	for {
		fields, err := reader.Read()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return fmt.Errorf("line %v: %v", nline, err)
		}
		select {
		case <-ctx.Done():
			return ErrCanceled
		default:
		}
		//removes csv header
		if nline == 0 && p.hasHeader {
			nline++
			continue
		}
		nline++
		//in each row, loop csv indexes to get data
		for _, idx := range p.indexes {
			if len(fields) <= idx {
				logger.Warnf("line %v: invalid index '%v'", nline, idx)
				break
			}
			data := strings.TrimSpace(fields[idx])

			// special cases
			if p.checks(xlist.IPv4) {
				ip, ipnet, err := net.ParseCIDR(data)
				if err == nil && ip.To4() != nil {
					out <- item{res: xlist.IPv4, format: xlistd.CIDR, name: ipnet.String()}
					nitems++
					if p.limited(nitems) {
						return nil
					}
					continue
				}
			}
			if p.checks(xlist.IPv6) {
				ip, ipnet, err := net.ParseCIDR(data)
				if err == nil && ip.To4() == nil {
					out <- item{res: xlist.IPv6, format: xlistd.CIDR, name: ipnet.String()}
					nitems++
					if p.limited(nitems) {
						return nil
					}
					continue
				}
			}
			//generic cases
			valid := false
			var rtype xlist.Resource
			for _, r := range p.resources {
				data, valid = xlist.Canonicalize(data, r)
				if valid {
					rtype = r
					break
				}
			}
			if !valid {
				logger.Warnf("line %v: not valid resource '%s'", nline, data)
				continue
			}
			//generic
			out <- item{res: rtype, format: xlistd.Plain, name: data}
			nitems++
			if p.limited(nitems) {
				return nil
			}
		}
	}
}

func (p csvConv) checks(r xlist.Resource) bool {
	return r.InArray(p.resources)
}

func (p csvConv) limited(nitems int) bool {
	return p.limit > 0 && nitems > p.limit
}
