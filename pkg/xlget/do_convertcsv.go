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

	"github.com/luids-io/core/utils/yalogi"
	"github.com/luids-io/core/xlist"
)

// CsvConv implementes a converter from csv format
type CsvConv struct {
	logger    yalogi.Logger
	Indexes   []int
	Resources []xlist.Resource
	Comma     rune
	Comment   rune
	HasHeader bool
	Limit     int
	Opts      ConvertOpts
}

// SetLogger implements Converter
func (p *CsvConv) SetLogger(l yalogi.Logger) {
	p.logger = l
}

// Convert implements Converter
func (p CsvConv) Convert(ctx context.Context, in io.Reader, out io.Writer) (map[xlist.Resource]int, error) {
	account := emptyAccount()
	nline := 0

	reader := csv.NewReader(bufio.NewReader(in))
	reader.Comma = p.Comma
	reader.Comment = p.Comment

	for {
		fields, err := reader.Read()
		if err == io.EOF {
			return account, nil
		} else if err != nil {
			return account, fmt.Errorf("line %v: %v", nline, err)
		}
		select {
		case <-ctx.Done():
			return account, ErrCanceled
		default:
		}
		//removes csv header
		if nline == 0 && p.HasHeader {
			nline++
			continue
		}
		nline++
		//in each row, loop csv indexes to get data
		for _, idx := range p.Indexes {
			if len(fields) <= idx {
				p.logger.Warnf("line %v: invalid index '%v'", nline, idx)
				break
			}
			data := strings.TrimSpace(fields[idx])

			// special cases
			if p.checks(xlist.IPv4) {
				ip, _, err := net.ParseCIDR(data)
				if err == nil && ip.To4() != nil {
					account[xlist.IPv4] = account[xlist.IPv4] + 1
					fmt.Fprintf(out, "ip4,cidr,%s\n", data)
					//check limits
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
					//check limits
					if p.Limit > 0 && nline > p.Limit {
						return account, nil
					}
					continue
				}
			}
			//generic cases
			valid := false
			var rtype xlist.Resource
		LOOPRESOURCES:
			for _, r := range p.Resources {
				if xlist.ValidResource(data, r) {
					valid = true
					rtype = r
					break LOOPRESOURCES
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
					//check limits
					if p.Limit > 0 && nline > p.Limit {
						return account, nil
					}
					continue
				}
			}
			//generic
			account[rtype] = account[rtype] + 1
			fmt.Fprintf(out, "%s,plain,%s\n", rtype, data)
		}
		//check limits
		if p.Limit > 0 && nline > p.Limit {
			return account, nil
		}
	}
}

func (p CsvConv) checks(r xlist.Resource) bool {
	if len(p.Resources) == 0 {
		return true
	}
	return r.InArray(p.Resources)
}
