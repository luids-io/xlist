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

// HostsConv implements a conversor from a /etc/hosts file format
type HostsConv struct {
	logger       yalogi.Logger
	Resources    []xlist.Resource
	WithDefaults bool
	Limit        int
	Opts         ConvertOpts
}

// SetLogger implements Converter interface
func (p *HostsConv) SetLogger(l yalogi.Logger) {
	p.logger = l
}

// Convert implements Converter interface
func (p HostsConv) Convert(ctx context.Context, in io.Reader, out io.Writer) (map[xlist.Resource]int, error) {
	account := emptyAccount()
	nline := 0
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return account, ErrCanceled
		default:
		}

		nline++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			_, isDefault := hostDefaults[fields[0]]
			if p.checks(xlist.IPv4) && xlist.ValidResource(fields[0], xlist.IPv4) {
				if !isDefault || (isDefault && p.WithDefaults) {
					account[xlist.IPv4] = account[xlist.IPv4] + 1
					fmt.Fprintf(out, "ip4,plain,%s\n", fields[0])
				}
			}
			if p.checks(xlist.IPv6) && xlist.ValidResource(fields[0], xlist.IPv6) {
				if !isDefault || (isDefault && p.WithDefaults) {
					account[xlist.IPv6] = account[xlist.IPv6] + 1
					fmt.Fprintf(out, "ip6,plain,%s\n", fields[0])
				}
			}
		}
		if len(fields) > 1 {
			added := false
			_, isDefault := hostDefaults[fields[1]]
			if p.checks(xlist.IPv4) && xlist.ValidResource(fields[1], xlist.IPv4) {
				if !isDefault || (isDefault && p.WithDefaults) {
					account[xlist.IPv4] = account[xlist.IPv4] + 1
					fmt.Fprintf(out, "ip4,plain,%s\n", fields[1])
					added = true
				}
			}
			if p.checks(xlist.IPv6) && xlist.ValidResource(fields[1], xlist.IPv6) && !added {
				if !isDefault || (isDefault && p.WithDefaults) {
					account[xlist.IPv6] = account[xlist.IPv6] + 1
					fmt.Fprintf(out, "ip6,plain,%s\n", fields[1])
					added = true
				}
			}
			if p.checks(xlist.Domain) && xlist.ValidResource(fields[1], xlist.Domain) && !added {
				if !isDefault || (isDefault && p.WithDefaults) {
					// apply opts
					if p.Opts.MinDomain > 0 {
						depth := len(strings.Split(fields[1], "."))
						if p.Opts.MinDomain > depth {
							account[xlist.Domain] = account[xlist.Domain] + 1
							fmt.Fprintf(out, "domain,sub,%s\n", fields[1])
						}
					} else {
						account[xlist.Domain] = account[xlist.Domain] + 1
						fmt.Fprintf(out, "domain,plain,%s\n", fields[1])
					}
				}
			}
		}
		if p.Limit > 0 && nline > p.Limit {
			return account, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return account, fmt.Errorf("scanning input: %v", err)
	}
	return account, nil
}

func (p HostsConv) checks(r xlist.Resource) bool {
	if len(p.Resources) == 0 {
		return true
	}
	return r.InArray(p.Resources)
}
