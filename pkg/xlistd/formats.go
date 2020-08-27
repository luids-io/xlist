// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlistd

import (
	"fmt"
	"strings"
)

// Format stores the type of available formats for lists.
type Format int

// List of formats.
const (
	Plain Format = iota
	CIDR
	Sub
)

func (f Format) string() string {
	switch f {
	case Plain:
		return "plain"
	case CIDR:
		return "cidr"
	case Sub:
		return "sub"
	default:
		return ""
	}
}

// String implements stringer interface.
func (f Format) String() string {
	s := f.string()
	if s == "" {
		return fmt.Sprintf("unkown(%d)", f)
	}
	return s
}

// ToFormat returns the format type from its string representation.
func ToFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "plain":
		return Plain, nil
	case "cidr":
		return CIDR, nil
	case "sub":
		return Sub, nil
	default:
		return Format(-1), fmt.Errorf("invalid format %s", s)
	}
}
