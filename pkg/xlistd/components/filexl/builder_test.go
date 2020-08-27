// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package filexl_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/apiservice"
	"github.com/luids-io/xlist/pkg/xlistd"
	"github.com/luids-io/xlist/pkg/xlistd/components/filexl"
)

var testdatabase1 = []xlistd.ListDef{
	{ID: "list1",
		Class: filexl.ComponentClass},
	{ID: "list2",
		Class:  filexl.ComponentClass,
		Source: "nonexistent"},
	{ID: "list3",
		Class:  filexl.ComponentClass,
		Source: "testfile1.xlist"},
	{ID: "list4",
		Class:     filexl.ComponentClass,
		Source:    "testfile1.xlist",
		Resources: []xlist.Resource{xlist.IPv4, xlist.Domain}},
	{ID: "list5",
		Class:     filexl.ComponentClass,
		Source:    "testfile1.xlist",
		Resources: xlist.Resources},
	{ID: "list6",
		Class:     filexl.ComponentClass,
		Source:    "testfile1.xlist",
		Resources: xlist.Resources,
		Opts:      map[string]interface{}{"reason": "aaaa"}},
	{ID: "list7",
		Class:     filexl.ComponentClass,
		Source:    "testfile1.xlist",
		Resources: xlist.Resources,
		Opts:      map[string]interface{}{"reason": 10}},
	{ID: "list8",
		Class:     filexl.ComponentClass,
		Source:    "testfile1.xlist",
		Resources: xlist.Resources,
		Opts:      map[string]interface{}{"autoreload": true}},
	{ID: "list9",
		Class:     filexl.ComponentClass,
		Source:    "testfile1.xlist",
		Resources: xlist.Resources,
		Opts:      map[string]interface{}{"autoreload": 10}},
	{ID: "list10",
		Class:     filexl.ComponentClass,
		Source:    "testfile1.xlist",
		Resources: xlist.Resources,
		Opts:      map[string]interface{}{"autoreload": true, "unsafereload": true}},
}

func TestBuild(t *testing.T) {
	tdir, err := filepath.Abs(testdir)
	if err != nil {
		t.Fatalf("invalid testdir %s", testdir)
	}
	b := xlistd.NewBuilder(apiservice.NewRegistry(), xlistd.DataDir(tdir))
	//define and do tests
	var tests = []struct {
		listid  string
		wantErr string
	}{
		{"list1", ""},
		{"list2", "doesn't exists"},
		{"list3", ""},
		{"list4", ""},
		{"list5", ""},
		{"list6", ""},
		{"list7", "invalid 'reason'"},
		{"list8", ""},
		{"list9", "invalid 'autoreload'"},
		{"list10", ""},
	}
	for _, test := range tests {
		def, ok := xlistd.FilterID(test.listid, testdatabase1)
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
