// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package responsewr_test

import (
	"strings"
	"testing"

	"github.com/luids-io/xlist/pkg/components/mockxl"
	"github.com/luids-io/xlist/pkg/wrappers/responsewr"

	"github.com/luids-io/core/xlist"
	listbuilder "github.com/luids-io/xlist/pkg/builder"
)

var (
	onlyIP     = []xlist.Resource{xlist.IPv4, xlist.IPv6}
	onlyIPv4   = []xlist.Resource{xlist.IPv4}
	onlyIPv6   = []xlist.Resource{xlist.IPv6}
	onlyDomain = []xlist.Resource{xlist.Domain}
)

var testdatabase1 = []listbuilder.ListDef{
	{ID: "list1",
		Class:     mockxl.BuildClass,
		Resources: onlyIPv4,
		Wrappers:  []listbuilder.WrapperDef{{Class: responsewr.BuildClass}}},
	{ID: "list2",
		Class:     mockxl.BuildClass,
		Resources: onlyIPv4,
		Wrappers: []listbuilder.WrapperDef{
			{Class: responsewr.BuildClass,
				Opts: map[string]interface{}{
					"negate":  true,
					"ttl":     10,
					"reason":  "razon",
					"preffix": "prefijo"}}}},
	{ID: "list3",
		Class:     mockxl.BuildClass,
		Resources: onlyIPv4,
		Wrappers: []listbuilder.WrapperDef{
			{Class: responsewr.BuildClass, Opts: map[string]interface{}{"preffixid": true}}}},
	{ID: "list4",
		Class:     mockxl.BuildClass,
		Resources: onlyIPv4,
		Wrappers: []listbuilder.WrapperDef{
			{Class: responsewr.BuildClass, Opts: map[string]interface{}{"ttl": "akk"}}}},
	{ID: "list5",
		Class:     mockxl.BuildClass,
		Resources: onlyIPv4,
		Wrappers: []listbuilder.WrapperDef{
			{Class: responsewr.BuildClass, Opts: map[string]interface{}{"negate": "akk"}}}},
	{ID: "list6",
		Class:     mockxl.BuildClass,
		Resources: onlyIPv4,
		Wrappers: []listbuilder.WrapperDef{
			{Class: responsewr.BuildClass, Opts: map[string]interface{}{"preffixid": "akk"}}}},
}

func TestBuild(t *testing.T) {
	builder := listbuilder.New()

	//define and do tests
	var tests = []struct {
		listid  string
		wantErr string
	}{
		{"list1", ""},
		{"list2", ""},
		{"list3", ""},
		{"list4", "invalid 'ttl'"},
		{"list5", "invalid 'negate'"},
		{"list6", "invalid 'preffixid'"},
	}
	for _, test := range tests {
		def, ok := listbuilder.FilterID(test.listid, testdatabase1)
		if !ok {
			t.Errorf("can't find id %s in database tests", test.listid)
			continue
		}
		_, err := builder.Build(def)
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
