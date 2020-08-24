// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package cachewr_test

import (
	"strings"
	"testing"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/apiservice"
	"github.com/luids-io/xlist/pkg/xlistd/builder"
	"github.com/luids-io/xlist/pkg/xlistd/components/mockxl"
	"github.com/luids-io/xlist/pkg/xlistd/wrappers/cachewr"
)

var (
	onlyIP     = []xlist.Resource{xlist.IPv4, xlist.IPv6}
	onlyIPv4   = []xlist.Resource{xlist.IPv4}
	onlyIPv6   = []xlist.Resource{xlist.IPv6}
	onlyDomain = []xlist.Resource{xlist.Domain}
)

var testdatabase1 = []builder.ListDef{
	{ID: "list1",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4,
		Wrappers:  []builder.WrapperDef{{Class: cachewr.WrapperClass}}},
	{ID: "list2",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4,
		Wrappers: []builder.WrapperDef{
			{Class: cachewr.WrapperClass,
				Opts: map[string]interface{}{"ttl": 10}}}},
	{ID: "list3",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4,
		Wrappers: []builder.WrapperDef{
			{Class: cachewr.WrapperClass,
				Opts: map[string]interface{}{"ttl": "aa"}}},
	},
	{ID: "list4",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4,
		Wrappers: []builder.WrapperDef{
			{Class: cachewr.WrapperClass,
				Opts: map[string]interface{}{"ttl": 0}}}},
	{ID: "list5",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4,
		Wrappers: []builder.WrapperDef{
			{Class: cachewr.WrapperClass,
				Opts: map[string]interface{}{"ttl": -1}}}},
	{ID: "list6",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4,
		Wrappers: []builder.WrapperDef{
			{Class: cachewr.WrapperClass,
				Opts: map[string]interface{}{"ttl": 10, "negativettl": 5}}}},
	{ID: "list7",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4,
		Wrappers: []builder.WrapperDef{
			{Class: cachewr.WrapperClass,
				Opts: map[string]interface{}{"ttl": 10, "negativettl": xlist.NeverCache}}}},
	{ID: "list8",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4,
		Wrappers: []builder.WrapperDef{
			{Class: cachewr.WrapperClass,
				Opts: map[string]interface{}{"ttl": 10, "negativettl": -2}}}},
}

func TestBuild(t *testing.T) {
	b := builder.New(apiservice.NewRegistry())

	//define and do tests
	var tests = []struct {
		listid  string
		wantErr string
	}{
		{"list1", ""},
		{"list2", ""},
		{"list3", "invalid 'ttl'"},
		{"list4", "invalid 'ttl'"},
		{"list5", "invalid 'ttl'"},
		{"list6", ""},
		{"list7", ""},
		{"list8", "invalid 'negativettl'"},
	}
	for _, test := range tests {
		def, ok := builder.FilterID(test.listid, testdatabase1)
		if !ok {
			t.Errorf("can't find id %s in database tests", test.listid)
			continue
		}
		_, err := b.Build(def)
		switch {
		case test.wantErr == "" && err == nil:
			//
		case test.wantErr == "" && err != nil:
			t.Errorf("unexpected error for %s: %v", test.listid, err)
		case test.wantErr != "" && err == nil:
			t.Errorf("expected error for %s", test.listid)
		case test.wantErr != "" && !strings.Contains(err.Error(), test.wantErr):
			t.Errorf("unexpected error for %s: %v", test.listid, err)
		}
	}
}
