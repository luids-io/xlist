// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlget

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/luids-io/core/xlist"
)

// Entry defines configuration entries format
type Entry struct {
	ID       string        `json:"id"`
	Update   Duration      `json:"update"`
	Sources  []EntrySource `json:"sources"`
	Disabled bool          `json:"disabled,omitempty"`
	Output   string        `json:"output,omitempty"`
	NoClean  bool          `json:"noclean,omitempty"`
	NoHash   bool          `json:"nohash,omitempty"`
}

// EntrySource defines configuration for sources
type EntrySource struct {
	URI         string           `json:"uri"`
	Format      string           `json:"format"`
	Compression string           `json:"compression,omitempty"`
	Filename    string           `json:"filename,omitempty"`
	Resources   []xlist.Resource `json:"resources,omitempty"`
	Limit       int              `json:"limit,omitempty"`
	FormatOpts  *FormatOpts      `json:"formatopts,omitempty"`
	ConvertOpts *ConvertOpts     `json:"convertopts,omitempty"`
}

// FormatOpts defines format options for conversors
type FormatOpts struct {
	Comma     string `json:"comma,omitempty"`
	Comment   string `json:"comment,omitempty"`
	HasHeader bool   `json:"header,omitempty"`
	Indexes   []int  `json:"indexes,omitempty"`
}

// ConvertOpts defines some conversion options for conversors
type ConvertOpts struct {
	MinDomain int `json:"mindomain,omitempty"`
}

// DefsFromFile returns an Entry slice of configuration
func DefsFromFile(path string) ([]Entry, error) {
	var entries []Entry
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %v", path, err)
	}
	byteValue, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %v", path, err)
	}
	err = json.Unmarshal(byteValue, &entries)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling downloads from json file %s: %v", path, err)
	}
	return entries, nil
}

// Validate checks if entry is valid
func (e Entry) Validate() error {
	if e.ID == "" {
		return errors.New("id can't be empty")
	}
	if len(e.Sources) == 0 {
		return errors.New("sources is required")
	}
	for _, source := range e.Sources {
		if source.URI == "" {
			return errors.New("uri is required")
		}
		if !ValidURI(source.URI) {
			return fmt.Errorf("invalid uri '%s'", source.URI)
		}
		if source.Format != "" {
			switch source.Format {
			case "csv":
			case "hosts":
			case "flat":
			case "xlist":
			default:
				return errors.New("unexpected format")
			}
		}
		if source.Compression != "" {
			switch source.Compression {
			case "none":
			case "gzip":
			case "zip":
			default:
				return errors.New("unexpected compression")
			}
		}
	}
	return nil
}

// Request returns a new request from a configuration entry
func (e Entry) Request() (Request, error) {
	request := Request{
		ID:      e.ID,
		Output:  e.Output,
		NoClean: e.NoClean,
		NoHash:  e.NoHash,
	}
	for _, src := range e.Sources {
		source := Source{
			URI:      src.URI,
			Filename: src.Filename,
		}
		switch src.Compression {
		case "gzip":
			source.Compression = Gzip
		case "zip":
			source.Compression = Zip
		}
		switch src.Format {
		case "csv":
			if len(src.Resources) == 0 {
				return Request{}, errors.New("resources is required for csv format")
			}
			if src.FormatOpts == nil {
				return Request{}, errors.New("formatopts is required for csv format")
			}
			if len(src.FormatOpts.Indexes) == 0 {
				return Request{}, errors.New("indexes is required for csv format")
			}
			csvconv := &CsvConv{
				Resources: src.Resources,
				Indexes:   src.FormatOpts.Indexes,
				Limit:     src.Limit,
				Comma:     ',',
				Comment:   '#',
				HasHeader: false,
			}
			csvconv.HasHeader = src.FormatOpts.HasHeader
			if src.FormatOpts.Comma != "" {
				runes := []rune(src.FormatOpts.Comma)
				if len(runes) > 0 {
					csvconv.Comma = runes[0]
				}
			}
			if src.FormatOpts.Comment != "" {
				runes := []rune(src.FormatOpts.Comment)
				if len(runes) > 0 {
					csvconv.Comment = runes[0]
				}
			}
			if src.ConvertOpts != nil {
				csvconv.Opts = *src.ConvertOpts
			}
			source.Converter = csvconv
		case "hosts":
			if len(src.Resources) == 0 {
				return Request{}, errors.New("resources is required for hosts format")
			}
			hostconv := &HostsConv{
				Resources: src.Resources,
				Limit:     src.Limit,
			}
			if src.ConvertOpts != nil {
				hostconv.Opts = *src.ConvertOpts
			}
			source.Converter = hostconv
		case "flat":
			if len(src.Resources) == 0 {
				return Request{}, errors.New("resources is required for flat format")
			}
			flatconv := &FlatConv{
				Resources: src.Resources,
				Limit:     src.Limit,
			}
			if src.ConvertOpts != nil {
				flatconv.Opts = *src.ConvertOpts
			}
			source.Converter = flatconv
		case "xlist":
			if len(src.Resources) == 0 {
				return Request{}, errors.New("resources is required for xlist format")
			}
			source.Converter = &XListConv{
				Resources: src.Resources,
				Limit:     src.Limit,
			}
		}
		request.Sources = append(request.Sources, source)
	}
	return request, nil
}

// Duration is used for unmarshalling durations. tip from: https://robreid.io/json-time-duration/
type Duration struct {
	time.Duration
}

// UnmarshalJSON implements interface
func (d *Duration) UnmarshalJSON(b []byte) (err error) {
	d.Duration, err = time.ParseDuration(strings.Trim(string(b), `"`))
	return
}

// MarshalJSON implements interface
func (d Duration) MarshalJSON() (b []byte, err error) {
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}
