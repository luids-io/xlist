// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package mockxl_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/components/mockxl"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

var testdatabase1 = []listbuilder.ListDef{
	{ID: "list1",
		Class: mockxl.BuildClass},
	{ID: "list2",
		Class:  mockxl.BuildClass,
		Source: "  true "},
	{ID: "list3",
		Class:  mockxl.BuildClass,
		Source: "false ,true , true"},
	{ID: "list4",
		Class:  mockxl.BuildClass,
		Source: "false,true,true,false"},
	{ID: "list5",
		Class:  mockxl.BuildClass,
		Source: "kk"},
	{ID: "list6",
		Class: mockxl.BuildClass,
		Opts:  map[string]interface{}{"ttl": "aa"}},
	{ID: "list7",
		Class: mockxl.BuildClass,
		Opts:  map[string]interface{}{"ttl": 10}},
	{ID: "list8",
		Class: mockxl.BuildClass,
		Opts:  map[string]interface{}{"fail": true}},
	{ID: "list9",
		Class: mockxl.BuildClass,
		Opts:  map[string]interface{}{"reason": "aaaa"}},
	{ID: "list10",
		Class: mockxl.BuildClass,
		Opts:  map[string]interface{}{"lazy": 10}},
}

func TestBuilderSources(t *testing.T) {
	builder := listbuilder.New()
	var tests = []struct {
		listid  string
		want    []bool
		wantErr bool
	}{
		{"list1", nil, false},
		{"list2", []bool{true}, false},
		{"list3", []bool{false, true, true}, false},
		{"list4", []bool{false, true, true, false}, false},
		{"list5", nil, true},
	}
	for _, test := range tests {
		def, _ := listbuilder.FilterID(test.listid, testdatabase1)
		checker, err := builder.Build(def)
		if test.wantErr && err == nil {
			t.Errorf("expected error for %s", test.listid)
		} else if test.wantErr && err != nil {
			continue
		} else if err != nil {
			t.Errorf("unexpected error for %s: %v", test.listid, err)
			continue
		}
		mock, ok := checker.(*mockxl.List)
		if !ok {
			t.Fatalf("can't cast mockup %s", test.listid)
		}
		if !cmpResults(mock.Results, test.want) {
			t.Errorf("results mismatch for %s: want=%v got=%v", test.listid, mock.Results, test.want)
		}
	}
}

func cmpResults(a, b []bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestBuilderProperties(t *testing.T) {
	builder := listbuilder.New()
	var tests = []struct {
		listid     string
		wantTTL    int
		wantReason string
		wantFail   bool
		wantLazy   time.Duration
		wantErr    bool
	}{
		{"list6", 0, "", false, 0, true},
		{"list7", 10, "", false, 0, false},
		{"list8", 0, "", true, 0, false},
		{"list9", 0, "aaaa", false, 0, false},
		{"list10", 0, "", false, 10 * time.Millisecond, false},
	}
	for _, test := range tests {
		def, _ := listbuilder.FilterID(test.listid, testdatabase1)
		checker, err := builder.Build(def)
		if test.wantErr && err == nil {
			t.Errorf("expected error for %s", test.listid)
		} else if test.wantErr && err != nil {
			continue
		} else if err != nil {
			t.Errorf("unexpectred error for %s: %v", test.listid, err)
			continue
		}
		mock, ok := checker.(*mockxl.List)
		if !ok {
			t.Fatalf("can't cast mockup %s", test.listid)
		}
		if mock.TTL != test.wantTTL {
			t.Errorf("ttl mismatch for %s: want=%v got=%v", test.listid, mock.TTL, test.wantTTL)
		}
		if mock.Reason != test.wantReason {
			t.Errorf("reason mismatch for %s: want=%v got=%v", test.listid, mock.Reason, test.wantReason)
		}
		if mock.Fail != test.wantFail {
			t.Errorf("fail mismatch for %s: want=%v got=%v", test.listid, mock.Fail, test.wantFail)
		}
		if mock.Lazy != test.wantLazy {
			t.Errorf("lazy mismatch for %s: want=%v got=%v", test.listid, mock.Lazy, test.wantLazy)
		}
	}
}

func ExampleBuilder() {
	// instande builder and register class
	builder := listbuilder.New()

	// create a definition for a list that checks ipv4
	// and returns true to all checks
	listdef1 := listbuilder.ListDef{
		ID:        "list1",
		Class:     "mock",
		Resources: []xlist.Resource{xlist.IPv4},
		Source:    "true",
	}
	rbl1, err := builder.Build(listdef1)
	if err != nil {
		log.Fatalln("this should not happen")
	}

	resp, err := rbl1.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != nil || !resp.Result {
		log.Fatalln("this should not happen")
	}
	fmt.Println("check ip4 rbl1:", resp.Result, resp.Reason)

	resp, err = rbl1.Check(context.Background(), "www.google.com", xlist.Domain)
	if err != xlist.ErrNotImplemented {
		log.Fatalln("this should not happen")
	}
	fmt.Println("check domain rbl1:", err)

	// create a definition for a list that checks domain
	// and fails
	listdef2 := listbuilder.ListDef{
		ID:        "list2",
		Class:     "mock",
		Resources: []xlist.Resource{xlist.IPv4, xlist.Domain},
		Opts: map[string]interface{}{
			"fail": true,
		},
	}

	rbl2, err := builder.Build(listdef2)
	if err != nil {
		log.Fatalln("this should not happen")
	}

	resp, err = rbl2.Check(context.Background(), "www.google.com", xlist.Domain)
	if err != xlist.ErrNotAvailable {
		log.Fatalln("this should not happen")
	}
	fmt.Println("check domain rbl2:", err)

	// Output:
	// check ip4 rbl1: true The resource is on the mockup list
	// check domain rbl1: not implemented
	// check domain rbl2: not available
}
