// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/luids-io/core/utils/yalogi"
	"github.com/luids-io/core/xlist"
)

//FlatConv implements a conversor from a flat file
type FlatConv struct {
	logger    yalogi.Logger
	Resources []xlist.Resource
	Limit     int
	Opts      ConvertOpts
}

// SetLogger implements Converter interface
func (p *FlatConv) SetLogger(l yalogi.Logger) {
	p.logger = l
}

// Convert implements Converter interface
func (p FlatConv) Convert(ctx context.Context, in io.Reader, out io.Writer) (map[xlist.Resource]int, error) {
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
		if p.checks(xlist.Domain) && strings.HasPrefix(data, "*.") {
			subdomain := strings.TrimPrefix(data, "*.")
			if !xlist.ValidResource(subdomain, xlist.Domain) {
				p.logger.Warnf("line %v: not valid resource '%s'", nline, data)
				continue
			}
			account[xlist.Domain] = account[xlist.Domain] + 1
			fmt.Fprintf(out, "domain,sub,%s\n", subdomain)
			// check limits
			if p.Limit > 0 && nline > p.Limit {
				return account, nil
			}
			continue
		}
		if p.checks(xlist.IPv4) {
			ip, _, err := net.ParseCIDR(data)
			if err == nil && ip.To4() != nil {
				account[xlist.IPv4] = account[xlist.IPv4] + 1
				fmt.Fprintf(out, "ip4,cidr,%s\n", data)
				// check limits
				if p.Limit > 0 && nline > p.Limit {
					return account, nil
				}
				continue
			}
		}
		if p.checks(xlist.IPv6) {
			ip, _, err := net.ParseCIDR(data)
			if err == nil && ip.To4() == nil {
				account[xlist.IPv6] = account[xlist.IPv6] + 1
				fmt.Fprintf(out, "ip6,cidr,%s\n", data)
				// check limits
				if p.Limit > 0 && nline > p.Limit {
					return account, nil
				}
				continue
			}
		}

		// generic cases
		valid := false
		var rtype xlist.Resource
		for _, r := range p.Resources {
			if xlist.ValidResource(data, r) {
				valid = true
				rtype = r
				break
			}
		}
		if !valid {
			p.logger.Warnf("line %v: not valid resource '%s'", nline, data)
			continue
		}
		// apply opts
		if rtype == xlist.Domain && p.Opts.MinDomain > 0 {
			depth := len(strings.Split(data, "."))
			if p.Opts.MinDomain > depth {
				account[rtype] = account[rtype] + 1
				fmt.Fprintf(out, "%s,sub,%s\n", rtype, data)
				// check limits
				if p.Limit > 0 && nline > p.Limit {
					return account, nil
				}
				continue
			}
		}
		account[rtype] = account[rtype] + 1
		fmt.Fprintf(out, "%s,plain,%s\n", rtype, data)

		// check limits
		if p.Limit > 0 && nline > p.Limit {
			return account, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return account, fmt.Errorf("scanning input: %v", err)
	}
	return account, nil
}

func (p FlatConv) checks(r xlist.Resource) bool {
	if len(p.Resources) == 0 {
		return true
	}
	return r.InArray(p.Resources)
}
