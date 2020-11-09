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

	"github.com/luids-io/api/xlist"
)

// Entry defines configuration entries format
type Entry struct {
	ID         string         `json:"id"`
	Disabled   bool           `json:"disabled,omitempty"`
	Update     Duration       `json:"update"`
	Sources    []Source       `json:"sources"`
	Transforms *TransformOpts `json:"transforms,omitempty"`
	NoClean    bool           `json:"noclean,omitempty"`
	NoHash     bool           `json:"nohash,omitempty"`
	Output     string         `json:"output,omitempty"`
}

// Copy returns a copy of Entry.
func (e Entry) Copy() (dst Entry) {
	dst.ID = e.ID
	dst.Disabled = e.Disabled
	dst.Update = e.Update
	if len(e.Sources) > 0 {
		dst.Sources = make([]Source, 0, len(e.Sources))
		for _, source := range e.Sources {
			dst.Sources = append(dst.Sources, source.Copy())
		}
	}
	if e.Transforms != nil {
		dst.Transforms = &TransformOpts{}
		*dst.Transforms = *e.Transforms
	}
	dst.NoClean = e.NoClean
	dst.Output = e.Output
	return
}

// Source defines configuration for sources
type Source struct {
	URI         string      `json:"uri"`
	Filename    string      `json:"filename,omitempty"`
	Compression Compression `json:"compression,omitempty"`

	Format     FormatSource     `json:"format"`
	FormatOpts *FormatOpts      `json:"formatopts,omitempty"`
	Resources  []xlist.Resource `json:"resources,omitempty"`
	Limit      int              `json:"limit,omitempty"`
}

// Copy returns a copy of Source.
func (s Source) Copy() (dst Source) {
	dst.URI = s.URI
	dst.Filename = s.Filename
	dst.Compression = s.Compression
	dst.Format = s.Format
	if s.FormatOpts != nil {
		dst.FormatOpts = &FormatOpts{}
		*dst.FormatOpts = *s.FormatOpts
		if len(s.FormatOpts.Indexes) > 0 {
			dst.FormatOpts.Indexes = make([]int, len(s.FormatOpts.Indexes), len(s.FormatOpts.Indexes))
			copy(dst.FormatOpts.Indexes, s.FormatOpts.Indexes)
		}
	}
	dst.Resources = make([]xlist.Resource, len(s.Resources), len(s.Resources))
	copy(dst.Resources, s.Resources)
	dst.Limit = s.Limit
	return
}

// FormatOpts defines format options for conversors
type FormatOpts struct {
	Comma      string `json:"comma,omitempty"`
	Comment    string `json:"comment,omitempty"`
	HasHeader  bool   `json:"header,omitempty"`
	Indexes    []int  `json:"indexes,omitempty"`
	LazyQuotes bool   `json:"lazyquotes,omitempty"`
}

// TransformOpts defines transformations.
type TransformOpts struct {
	TLDPlusOne bool `json:"tldplusone,omitempty"`
}

// Compression defines compression algs
type Compression int

// List of compression values
const (
	None Compression = iota
	Gzip
	Zip
)

// Duration is used for unmarshalling durations. tip from: https://robreid.io/json-time-duration/
type Duration struct {
	time.Duration
}

func (c Compression) string() string {
	switch c {
	case None:
		return "none"
	case Gzip:
		return "gzip"
	case Zip:
		return "zip"
	default:
		return ""
	}
}

func (c Compression) String() string {
	s := c.string()
	if s == "" {
		return fmt.Sprintf("unkown(%d)", c)
	}
	return s
}

// MarshalJSON implements interface for struct marshalling.
func (c Compression) MarshalJSON() ([]byte, error) {
	s := c.string()
	if s == "" {
		return nil, fmt.Errorf("invalid value %v for compression", c)
	}
	return json.Marshal(s)
}

// UnmarshalJSON implements interface for struct unmarshalling.
func (c *Compression) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	case "none":
		*c = None
		return nil
	case "gzip":
		*c = Gzip
		return nil
	case "zip":
		*c = Zip
		return nil
	default:
		return fmt.Errorf("cannot unmarshal compression %s", s)
	}
}

// FormatSource defines source formats
type FormatSource int

// List of source format values
const (
	XList FormatSource = iota
	Flat
	CSV
	Hosts
)

func (f FormatSource) string() string {
	switch f {
	case XList:
		return "xlist"
	case Flat:
		return "flat"
	case CSV:
		return "csv"
	case Hosts:
		return "hosts"
	default:
		return ""
	}
}

func (f FormatSource) String() string {
	s := f.string()
	if s == "" {
		return fmt.Sprintf("unkown(%d)", f)
	}
	return s
}

// MarshalJSON implements interface for struct marshalling.
func (f FormatSource) MarshalJSON() ([]byte, error) {
	s := f.string()
	if s == "" {
		return nil, fmt.Errorf("invalid value %v for source format", f)
	}
	return json.Marshal(s)
}

// UnmarshalJSON implements interface for struct unmarshalling.
func (f *FormatSource) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	case "xlist":
		*f = XList
		return nil
	case "flat":
		*f = Flat
		return nil
	case "csv":
		*f = CSV
		return nil
	case "hosts":
		*f = Hosts
		return nil
	default:
		return fmt.Errorf("cannot unmarshal compression %s", s)
	}
}

// EntryDefsFromFile returns an Entry slice of configuration
func EntryDefsFromFile(path string) ([]Entry, error) {
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
		return nil, fmt.Errorf("unmarshalling from json file %s: %v", path, err)
	}
	return entries, nil
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

// ValidateEntry checks if entry is valid
func ValidateEntry(e Entry) error {
	if e.ID == "" {
		return errors.New("id can't be empty")
	}
	if e.Update.Duration == 0 {
		return errors.New("update is required")
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
		if !ValidCompression(source.Compression) {
			return fmt.Errorf("invalid compression '%s'", source.Compression)
		}
		if !ValidFormatSource(source.Format) {
			return fmt.Errorf("invalid source '%s'", source.Format)
		}
	}
	return nil
}
