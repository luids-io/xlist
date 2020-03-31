// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/luids-io/core/utils/yalogi"
	"github.com/luids-io/core/xlist"
)

// XListConv implements a converter from an xlist format
type XListConv struct {
	logger    yalogi.Logger
	Resources []xlist.Resource
	Limit     int
}

// SetLogger implements Converter interface
func (p *XListConv) SetLogger(l yalogi.Logger) {
	p.logger = l
}

// Convert implements Converter interface
func (p XListConv) Convert(ctx context.Context, in io.Reader, out io.Writer) (map[xlist.Resource]int, error) {
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
		fields := strings.Split(line, ",")
		if len(fields) < 3 {
			return account, fmt.Errorf("line %v: invalid format", nline)
		}
		resource, err := xlist.ToResource(fields[0])
		if err != nil {
			return account, fmt.Errorf("line %v: not valid resource", nline)
		}
		if !p.checks(resource) {
			continue
		}
		account[resource] = account[resource] + 1

		fmt.Fprintf(out, "%s,%s,%s\n", fields[0], fields[1], fields[2])
		if p.Limit > 0 && nline > p.Limit {
			return account, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return account, fmt.Errorf("scanning input: %v", err)
	}
	return account, nil
}

func (p XListConv) checks(r xlist.Resource) bool {
	if len(p.Resources) == 0 {
		return true
	}
	return r.InArray(p.Resources)
}
