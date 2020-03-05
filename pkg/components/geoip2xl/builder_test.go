// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package geoip2xl_test

import (
	"strings"
	"testing"

	"github.com/luids-io/xlist/pkg/components/geoip2xl"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

var testdatabase1 = []listbuilder.ListDef{
	{ID: "list1",
		Class: geoip2xl.BuildClass},
	{ID: "list2",
		Class:  geoip2xl.BuildClass,
		Source: "nonexistent"},
	{ID: "list3",
		Class:  geoip2xl.BuildClass,
		Source: testdb1},
	{ID: "list4",
		Class:     geoip2xl.BuildClass,
		Source:    testdb1,
		Resources: []xlist.Resource{xlist.IPv4, xlist.Domain}},
	{ID: "list5",
		Class:     geoip2xl.BuildClass,
		Source:    testdb1,
		Resources: []xlist.Resource{xlist.IPv4}},
	{ID: "list6",
		Class:     geoip2xl.BuildClass,
		Source:    testdb1,
		Resources: []xlist.Resource{xlist.IPv4},
		Opts:      map[string]interface{}{"reason": "aaaa"}},
	{ID: "list7",
		Class:     geoip2xl.BuildClass,
		Source:    testdb1,
		Resources: []xlist.Resource{xlist.IPv4},
		Opts:      map[string]interface{}{"reason": 10}},
	{ID: "list8",
		Class:     geoip2xl.BuildClass,
		Source:    testdb1,
		Resources: []xlist.Resource{xlist.IPv4},
		Opts:      map[string]interface{}{"countries": []string{"es", "gb"}}},
	{ID: "list9",
		Class:     geoip2xl.BuildClass,
		Source:    testdb1,
		Resources: []xlist.Resource{xlist.IPv4},
		Opts:      map[string]interface{}{"countries": []interface{}{"es", "gb"}}},
	{ID: "list10",
		Class:     geoip2xl.BuildClass,
		Source:    testdb1,
		Resources: []xlist.Resource{xlist.IPv4},
		Opts:      map[string]interface{}{"countries": []interface{}{"es", 5}}},
	{ID: "list11",
		Class:     geoip2xl.BuildClass,
		Source:    testdb1,
		Resources: []xlist.Resource{xlist.IPv4},
		Opts:      map[string]interface{}{"countries": []string{"es", "gb"}, "reverse": true}},
	{ID: "list12",
		Class:     geoip2xl.BuildClass,
		Source:    testdb1,
		Resources: []xlist.Resource{xlist.IPv4},
		Opts:      map[string]interface{}{"countries": []string{"es", "gb"}, "reverse": 5}},
}

func TestBuild(t *testing.T) {
	builder := listbuilder.New()
	//define and do tests
	var tests = []struct {
		listid  string
		wantErr string
	}{
		{"list1", "'source' is required"},
		{"list2", "doesn't exists"},
		{"list3", "invalid 'resources'"},
		{"list4", "invalid 'resources'"},
		{"list5", ""},
		{"list6", ""},
		{"list7", "invalid 'reason'"},
		{"list8", ""},
		{"list9", ""},
		{"list10", "invalid 'countries'"},
		{"list11", ""},
		{"list12", "invalid 'reverse'"},
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
