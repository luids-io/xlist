// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package parallelxl_test

import (
	"strings"
	"testing"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/apiservice"
	"github.com/luids-io/xlist/pkg/xlistd"
	"github.com/luids-io/xlist/pkg/xlistd/components/mockxl"
	"github.com/luids-io/xlist/pkg/xlistd/components/parallelxl"
)

var (
	onlyIP     = []xlist.Resource{xlist.IPv4, xlist.IPv6}
	onlyIPv4   = []xlist.Resource{xlist.IPv4}
	onlyIPv6   = []xlist.Resource{xlist.IPv6}
	onlyDomain = []xlist.Resource{xlist.Domain}
)

var testmocks = []xlistd.ListDef{
	{ID: "mock1",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4},
	{ID: "mock2",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4,
		Source:    "true",
		Opts:      map[string]interface{}{"reason": "mock2"}},
	{ID: "mock3",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4,
		Opts:      map[string]interface{}{"lazy": 10}},
	{ID: "mock4",
		Class:     mockxl.ComponentClass,
		Resources: onlyIPv4,
		Source:    "true",
		Opts:      map[string]interface{}{"lazy": 10, "reason": "mock4"}},
	{ID: "mock5",
		Class:     mockxl.ComponentClass,
		Resources: onlyIP},
	{ID: "mock6",
		Class:     mockxl.ComponentClass,
		Resources: onlyDomain},
}

var testparallel1 = []xlistd.ListDef{
	{ID: "list1",
		Class:     parallelxl.ComponentClass,
		Resources: onlyIPv4,
		Contains:  []xlistd.ListDef{{ID: "mock1"}}},
	{ID: "list2",
		Class:     parallelxl.ComponentClass,
		Resources: onlyIPv4,
		Contains:  []xlistd.ListDef{{ID: "mock1"}, {ID: "mock2"}}},
	{ID: "list3",
		Class: parallelxl.ComponentClass},
	{ID: "list4",
		Class:     parallelxl.ComponentClass,
		Resources: onlyIP,
		Contains:  []xlistd.ListDef{{ID: "mock1"}, {ID: "mock2"}}},
	{ID: "list5",
		Class:     parallelxl.ComponentClass,
		Resources: onlyIP,
		Contains:  []xlistd.ListDef{{ID: "mock1"}, {ID: "mock5"}}},
	{ID: "list6",
		Class:     parallelxl.ComponentClass,
		Resources: onlyIPv4,
		Opts:      map[string]interface{}{"reason": 10},
		Contains:  []xlistd.ListDef{{ID: "mock1"}}},
	{ID: "list7",
		Class:     parallelxl.ComponentClass,
		Resources: onlyIPv4,
		Opts:      map[string]interface{}{"reason": "hey", "stoponerror": true},
		Contains:  []xlistd.ListDef{{ID: "mock1"}}},
	{ID: "list8",
		Class:     parallelxl.ComponentClass,
		Resources: onlyIPv4,
		Opts:      map[string]interface{}{"reason": "hey", "stoponerror": true},
		Contains:  []xlistd.ListDef{{ID: "mock1"}, {ID: "mock6"}}},
}

func TestBuild(t *testing.T) {
	b := xlistd.NewBuilder(apiservice.NewRegistry())

	//create mocks
	for _, defmock := range testmocks {
		_, err := b.Build(defmock)
		if err != nil {
			t.Fatalf("building mock %s: %v", defmock.ID, err)
		}
	}
	//define and do tests
	var tests = []struct {
		listid  string
		wantErr string
	}{
		{"list1", ""},
		{"list2", ""},
		{"list3", ""},
		{"list4", "doesn't checks resource"},
		{"list5", "doesn't checks resource"},
		{"list6", "reason"},
		{"list7", ""},
		{"list8", "doesn't checks resource"},
	}
	for _, test := range tests {
		def, _ := xlistd.FilterID(test.listid, testparallel1)
		_, err := b.Build(def)
		switch {
		case test.wantErr == "" && err == nil:
			//
		case test.wantErr == "" && err != nil:
			t.Errorf("unexpected error for %s: %v", test.listid, err)
		case test.wantErr != "" && err == nil:
			t.Errorf("expected error for %s", test.listid)
		case test.wantErr != "" && !strings.Contains(err.Error(), test.wantErr):
			t.Errorf("unexpectred error for %s: %v", test.listid, err)
		}
	}
}
