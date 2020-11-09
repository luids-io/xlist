// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/yalogi"
	"github.com/luids-io/xlist/pkg/xlistd"
)

type flatConv struct {
	resources []xlist.Resource
	limit     int
}

func (c flatConv) convert(ctx context.Context, in io.Reader, out chan<- item, logger yalogi.Logger) error {
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
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		//now removes after first blank...
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		data := strings.TrimSpace(fields[0])

		// special cases
		if c.checks(xlist.Domain) && strings.HasPrefix(data, "*.") {
			valid := false
			subdomain := strings.TrimPrefix(data, "*.")
			subdomain, valid = xlist.Canonicalize(subdomain, xlist.Domain)
			if !valid {
				logger.Warnf("line %v: not valid resource '%s'", nline, data)
				continue
			}
			out <- item{res: xlist.Domain, format: xlistd.Sub, name: subdomain}
			nitems++
			if c.limited(nitems) {
				return nil
			}
			continue
		}
		if c.checks(xlist.IPv4) {
			ip, ipnet, err := net.ParseCIDR(data)
			if err == nil && ip.To4() != nil {
				out <- item{res: xlist.IPv4, format: xlistd.CIDR, name: ipnet.String()}
				nitems++
				if c.limited(nitems) {
					return nil
				}
				continue
			}
		}
		if c.checks(xlist.IPv6) {
			ip, ipnet, err := net.ParseCIDR(data)
			if err == nil && ip.To4() == nil {
				out <- item{res: xlist.IPv6, format: xlistd.CIDR, name: ipnet.String()}
				nitems++
				if c.limited(nitems) {
					return nil
				}
				continue
			}
		}

		// generic cases
		valid := false
		var rtype xlist.Resource
		for _, r := range c.resources {
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
		out <- item{res: rtype, format: xlistd.Plain, name: data}
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

func (c flatConv) checks(r xlist.Resource) bool {
	return r.InArray(c.resources)
}

func (c flatConv) limited(nitems int) bool {
	return c.limit > 0 && nitems > c.limit
}
