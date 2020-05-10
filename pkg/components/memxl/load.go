// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package memxl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/luids-io/api/xlist"
)

// Data is used for bulk insertions
type Data struct {
	Resource xlist.Resource
	Format   xlist.Format
	Value    string
}

func (i Data) String() string {
	return fmt.Sprintf("%v,%v,%s", i.Resource, i.Format, i.Value)
}

// LoadData loads list from a data array
func (l *List) LoadData(ctx context.Context, data []Data) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for idx, item := range data {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := l.add(item.Resource, item.Format, item.Value)
		if err != nil {
			return fmt.Errorf("idx %v: invalid '%v'", idx, item)
		}
	}
	return nil
}

//LoadReader data to the list from an io.Reader with the memxl format
func (l *List) LoadReader(ctx context.Context, in io.Reader) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	//process file
	nline := 0
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		nline++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) < 3 {
			return fmt.Errorf("line %v: invalid line", nline)
		}
		resource, err := xlist.ToResource(fields[0])
		if err != nil {
			return fmt.Errorf("line %v: invalid resource type '%s'", nline, fields[0])
		}
		format, err := xlist.ToFormat(fields[1])
		if err != nil {
			return fmt.Errorf("line %v: invalid format type '%s'", nline, fields[1])
		}
		value := fields[2]
		if l.checks(resource) {
			err := l.add(resource, format, value)
			if err != nil {
				return fmt.Errorf("line %v: invalid '%v,%v,%s'", nline, resource, format, value)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanning input: %v", err)
	}
	return nil
}

// LoadFromData loads a hashmem list from a data array
func LoadFromData(list *List, data []Data, clearBefore bool) error {
	if clearBefore {
		list.Clear(context.Background())
	}
	return list.LoadData(context.Background(), data)
}

// LoadFromFile loads a hashmem list from a file content
func LoadFromFile(list *List, filename string, clearBefore bool) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("opening file: %v", err)
	}
	defer file.Close()

	if clearBefore {
		list.Clear(context.Background())
	}
	return list.LoadReader(context.Background(), file)
}
