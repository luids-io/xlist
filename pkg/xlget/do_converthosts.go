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

var hostDefaults = map[string]bool{
	"127.0.0.1":             true,
	"127.0.1.1":             true,
	"0.0.0.0":               true,
	"255.255.255.255":       true,
	"::1":                   true,
	"ff00::0":               true,
	"fe80::1%lo0":           true,
	"ff02::1":               true,
	"ff02::2":               true,
	"ff02::3":               true,
	"local":                 true,
	"localhost":             true,
	"localhost.localdomain": true,
	"broadcasthost":         true,
	"ip6-localhost":         true,
	"ip6-loopback":          true,
	"ip6-localnet":          true,
	"ip6-mcastprefix":       true,
	"ip6-allnodes":          true,
	"ip6-allrouters":        true,
	"ip6-allhosts":          true,
}

type hostsConv struct {
	resources    []xlist.Resource
	WithDefaults bool
	limit        int
}

func (h hostsConv) convert(ctx context.Context, in io.Reader, out chan<- item, logger yalogi.Logger) error {
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
		fields := strings.Fields(line)
		if len(fields) > 1 && h.checks(xlist.IPv4) {
			name, valid := xlist.Canonicalize(fields[1], xlist.IPv4)
			if valid {
				_, isDefault := hostDefaults[name]
				if !isDefault || (isDefault && h.WithDefaults) {
					out <- item{res: xlist.IPv4, format: xlistd.Plain, name: name}
					nitems++
					if h.limited(nitems) {
						return nil
					}
					continue
				}
			}
		}
		if len(fields) > 1 && h.checks(xlist.Domain) {
			name, valid := xlist.Canonicalize(fields[1], xlist.Domain)
			if valid {
				_, isDefault := hostDefaults[name]
				if !isDefault || (isDefault && h.WithDefaults) {
					out <- item{res: xlist.Domain, format: xlistd.Plain, name: name}
					nitems++
					if h.limited(nitems) {
						return nil
					}
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanning input: %v", err)
	}
	return nil
}

func (h hostsConv) checks(r xlist.Resource) bool {
	return r.InArray(h.resources)
}

func (h hostsConv) limited(nitems int) bool {
	return h.limit > 0 && nitems > h.limit
}
