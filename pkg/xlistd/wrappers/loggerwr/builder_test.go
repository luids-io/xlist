// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package loggerwr_test

import (
	"strings"
	"testing"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/apiservice"
	"github.com/luids-io/xlist/pkg/xlistd/builder"
	"github.com/luids-io/xlist/pkg/xlistd/components/mockxl"
	"github.com/luids-io/xlist/pkg/xlistd/wrappers/loggerwr"
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
		Wrappers:  []builder.WrapperDef{{Class: "logger"}}},
	{ID: "list2",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4,
		Wrappers: []builder.WrapperDef{
			{Class: loggerwr.WrapperClass,
				Opts: map[string]interface{}{"showpeer": true}}}},
	{ID: "list3",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4,
		Wrappers: []builder.WrapperDef{
			{Class: loggerwr.WrapperClass,
				Opts: map[string]interface{}{"showpeer": "aa"}}}},
	{ID: "list4",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4,
		Wrappers: []builder.WrapperDef{
			{Class: loggerwr.WrapperClass,
				Opts: map[string]interface{}{
					"found":    "warn",
					"notfound": "disable",
					"error":    "error"}}}},
	{ID: "list5",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4,
		Wrappers: []builder.WrapperDef{
			{Class: loggerwr.WrapperClass,
				Opts: map[string]interface{}{
					"found":    5,
					"notfound": "disable"}}}},
	{ID: "list6",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4,
		Wrappers: []builder.WrapperDef{
			{Class: loggerwr.WrapperClass,
				Opts: map[string]interface{}{
					"found":    "warn",
					"notfound": "noexiste"}}}},
}

func TestBuild(t *testing.T) {
	output := &logmockup{}
	b := builder.New(apiservice.NewRegistry(), builder.SetLogger(output))

	//define and do tests
	var tests = []struct {
		listid  string
		wantErr string
	}{
		{"list1", ""},
		{"list2", ""},
		{"list3", "invalid 'showpeer'"},
		{"list4", ""},
		{"list5", "invalid 'found'"},
		{"list6", "invalid 'notfound'"},
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
