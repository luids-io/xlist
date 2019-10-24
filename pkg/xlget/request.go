// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"errors"
	"fmt"
)

// Request defines request fields
type Request struct {
	ID      string
	Sources []Source
	Output  string
	NoClean bool
	NoHash  bool
}

// Source stores information about sources
type Source struct {
	URI         string
	Converter   Converter
	Filename    string
	Compression Compression
}

// Compression defines compression algs
type Compression int

// List of compression values
const (
	None Compression = iota
	Gzip
	Zip
)

func (r Request) validate() error {
	if r.ID == "" {
		return errors.New("ListID is required")
	}
	if len(r.Sources) == 0 {
		return errors.New("sources is required")
	}
	for _, source := range r.Sources {
		if source.URI == "" {
			return errors.New("uri is required")
		}
		if !ValidURI(source.URI) {
			return fmt.Errorf("invalid uri '%s'", source.URI)
		}
	}
	return nil
}
