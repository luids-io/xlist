// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package parallelxl_test

import (
	"strings"
	"testing"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/apiservice"
	"github.com/luids-io/xlist/pkg/builder"
	"github.com/luids-io/xlist/pkg/components/mockxl"
	"github.com/luids-io/xlist/pkg/components/parallelxl"
)

var (
	onlyIP     = []xlist.Resource{xlist.IPv4, xlist.IPv6}
	onlyIPv4   = []xlist.Resource{xlist.IPv4}
	onlyIPv6   = []xlist.Resource{xlist.IPv6}
	onlyDomain = []xlist.Resource{xlist.Domain}
)

var testmocks = []builder.ListDef{
	{ID: "mock1",
		Class:     mockxl.BuildClass,
		Resources: onlyIPv4},
	{ID: "mock2",
		Class:     mockxl.BuildClass,
		Resources: onlyIPv4,
		Source:    "true",
		Opts:      map[string]interface{}{"reason": "mock2"}},
	{ID: "mock3",
		Class:     mockxl.BuildClass,
		Resources: onlyIPv4,
		Opts:      map[string]interface{}{"lazy": 10}},
	{ID: "mock4",
		Class:     mockxl.BuildClass,
		Resources: onlyIPv4,
		Source:    "true",
		Opts:      map[string]interface{}{"lazy": 10, "reason": "mock4"}},
	{ID: "mock5",
		Class:     mockxl.BuildClass,
		Resources: onlyIP},
	{ID: "mock6",
		Class:     mockxl.BuildClass,
		Resources: onlyDomain},
}

var testparallel1 = []builder.ListDef{
	{ID: "list1",
		Class:     parallelxl.BuildClass,
		Resources: onlyIPv4,
		Contains:  []builder.ListDef{{ID: "mock1"}}},
	{ID: "list2",
		Class:     parallelxl.BuildClass,
		Resources: onlyIPv4,
		Contains:  []builder.ListDef{{ID: "mock1"}, {ID: "mock2"}}},
	{ID: "list3",
		Class: parallelxl.BuildClass},
	{ID: "list4",
		Class:     parallelxl.BuildClass,
		Resources: onlyIP,
		Contains:  []builder.ListDef{{ID: "mock1"}, {ID: "mock2"}}},
	{ID: "list5",
		Class:     parallelxl.BuildClass,
		Resources: onlyIP,
		Contains:  []builder.ListDef{{ID: "mock1"}, {ID: "mock5"}}},
	{ID: "list6",
		Class:     parallelxl.BuildClass,
		Resources: onlyIPv4,
		Opts:      map[string]interface{}{"reason": 10},
		Contains:  []builder.ListDef{{ID: "mock1"}}},
	{ID: "list7",
		Class:     parallelxl.BuildClass,
		Resources: onlyIPv4,
		Opts:      map[string]interface{}{"reason": "hey", "stoponerror": true},
		Contains:  []builder.ListDef{{ID: "mock1"}}},
	{ID: "list8",
		Class:     parallelxl.BuildClass,
		Resources: onlyIPv4,
		Opts:      map[string]interface{}{"reason": "hey", "stoponerror": true},
		Contains:  []builder.ListDef{{ID: "mock1"}, {ID: "mock6"}}},
}

func TestBuild(t *testing.T) {
	b := builder.New(apiservice.NewRegistry())

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
		def, _ := builder.FilterID(test.listid, testparallel1)
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
