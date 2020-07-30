// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlistd

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Category type is used for RBL classification.
type Category int

// Constants used for categories allowed
const (
	Blacklist Category = iota //blacklist
	Whitelist                 //whitelist
	Mixedlist                 //mixed
	Infolist                  //information
)

// Categories is a vector that constains all category types
var Categories = []Category{Blacklist, Whitelist, Mixedlist, Infolist}

// IsValid returns true if resource code is a valid
func (c Category) IsValid() bool {
	v := int(c)
	if v >= int(Blacklist) && v <= int(Infolist) {
		return true
	}
	return false
}

// String implements interface
func (c Category) String() string {
	switch c {
	case Blacklist:
		return "blacklist"
	case Whitelist:
		return "whitelist"
	case Mixedlist:
		return "mixedlist"
	case Infolist:
		return "infolist"
	}
	return fmt.Sprintf("unkown(%d)", c)
}

// MarshalJSON implements interface for struct marshalling
func (c Category) MarshalJSON() ([]byte, error) {
	s := ""
	switch c {
	case Blacklist:
		s = "blacklist"
	case Whitelist:
		s = "whitelist"
	case Mixedlist:
		s = "mixedlist"
	case Infolist:
		s = "infolist"
	default:
		return nil, fmt.Errorf("invalid value '%v' for category", s)
	}
	return json.Marshal(s)
}

// UnmarshalJSON implements interface for struct unmarshalling
func (c *Category) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "blacklist":
		*c = Blacklist
		return nil
	case "whitelist":
		*c = Whitelist
		return nil
	case "mixedlist":
		*c = Mixedlist
		return nil
	case "infolist":
		*c = Infolist
		return nil
	default:
		return fmt.Errorf("cannot unmarshal category '%s'", c)
	}
}

// ToCategory returns the category from its string representation
func ToCategory(s string) (Category, error) {
	switch strings.ToLower(s) {
	case "blacklist":
		return Blacklist, nil
	case "whitelist":
		return Whitelist, nil
	case "mixedlist":
		return Mixedlist, nil
	case "infolist":
		return Infolist, nil
	default:
		return Category(-1), fmt.Errorf("invalid category %s", s)
	}
}
