// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package builder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/luids-io/core/xlist"
)

// ListDef stores metadata info about RBL services.
// It's used by Builder type to build generic Check interfaces.
type ListDef struct {
	// ID must exist and be unique in databases for its correct operation
	ID string `json:"id"`
	// Class stores the component type of the RBL
	Class string `json:"class"`
	// Disabled flag
	Disabled bool `json:"disabled"`
	// Name or description of the list
	Name string `json:"name,omitempty"`
	// Category of the list
	Category xlist.Category `json:"category"`
	// Tags associated with the list
	Tags []string `json:"tags,omitempty"`
	// Resources is a list of the recource types supported
	Resources []xlist.Resource `json:"resources,omitempty"`
	// Web provides the website of the RBL
	Web string `json:"web,omitempty"`
	// Source provides the origin of the RBL
	Source string `json:"source,omitempty"`
	// TLS defines the configurationn of client protocol if it's supported by
	// the RBL
	TLS *ConfigTLS `json:"tls,omitempty"`
	// Wrappers definition of the list
	Wrappers []WrapperDef `json:"wrappers,omitempty"`
	// Contains stores child RBLs
	Contains []ListDef `json:"contains,omitempty"`
	// Opts custom options of the RBL
	Opts map[string]interface{} `json:"opts,omitempty"`
}

// WrapperDef stores metadata info about wrappers. Wrappers are used for provide
// additional funtionality to RBLs.
type WrapperDef struct {
	// Class stores the component type of the Wrapper
	Class string `json:"class"`
	// Opts custom options of the wrapper
	Opts map[string]interface{} `json:"opts,omitempty"`
}

//ConfigTLS stores information used in TLS connections
type ConfigTLS struct {
	CertFile     string `json:"certfile,omitempty"`
	KeyFile      string `json:"keyfile,omitempty"`
	ServerName   string `json:"servername,omitempty"`
	ServerCert   string `json:"servercert,omitempty"`
	CACert       string `json:"cacert,omitempty"`
	UseSystemCAs bool   `json:"systemca"`
}

// Filter* functions are useful for working with ListDef databases.
// All this functions uses linear search for do its job. Because that,
// this functions should not be used in critical paths.

// FilterID returns the listdef with the id in the ListDef slice.
func FilterID(id string, l []ListDef) (ListDef, bool) {
	for _, entry := range l {
		if entry.ID == id {
			return entry, true
		}
	}
	return ListDef{}, false
}

// FilterResource returns all listdefs that provides support for the resource
// passed.
func FilterResource(r xlist.Resource, l []ListDef) []ListDef {
	result := make([]ListDef, 0)
	for _, entry := range l {
	LOOPRES:
		for _, c := range entry.Resources {
			if c == r {
				result = append(result, entry)
				break LOOPRES
			}
		}
	}
	return result
}

// FilterClass returns all listdefs that correspond to the class passed.
func FilterClass(c string, l []ListDef) []ListDef {
	result := make([]ListDef, 0)
	for _, entry := range l {
		if entry.Class == c {
			result = append(result, entry)
		}
	}
	return result
}

// FilterCategory returns all listdefs of the category.
func FilterCategory(c xlist.Category, l []ListDef) []ListDef {
	result := make([]ListDef, 0)
	for _, entry := range l {
		if entry.Category == c {
			result = append(result, entry)
		}
	}
	return result
}

// FilterTag returns all listdefs that contains the tag. If tag is empty
// it will return lists with no tags.
func FilterTag(tag string, l []ListDef) []ListDef {
	result := make([]ListDef, 0)
	for _, entry := range l {
		if tag == "" {
			if entry.Tags == nil || len(entry.Tags) == 0 {
				result = append(result, entry)
				continue
			}
		}
	LOOPTAGS:
		for _, t := range entry.Tags {
			if tag == t {
				result = append(result, entry)
				break LOOPTAGS
			}
		}
	}
	return result
}

// ListDefsByID implements sort.Interface based on the ID field
type ListDefsByID []ListDef

func (a ListDefsByID) Len() int           { return len(a) }
func (a ListDefsByID) Less(i, j int) bool { return a[i].ID < a[j].ID }
func (a ListDefsByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// ListDefsByName implements sort.Interface based on the Name field
type ListDefsByName []ListDef

func (a ListDefsByName) Len() int { return len(a) }
func (a ListDefsByName) Less(i, j int) bool {
	return strings.ToUpper(a[i].Name) < strings.ToUpper(a[j].Name)
}
func (a ListDefsByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// DefsFromFile creates a slice of ListDef from a file
// in json format.
func DefsFromFile(path string) ([]ListDef, error) {
	var lists []ListDef
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, fmt.Errorf("opening file '%s': %v", path, err)
	}
	byteValue, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading file '%s': %v", path, err)
	}
	err = json.Unmarshal(byteValue, &lists)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling lists from json file '%s': %v", path, err)
	}
	return lists, nil
}
