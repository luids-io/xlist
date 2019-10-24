// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package builder_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/luids-io/core/xlist"
	listbuilder "github.com/luids-io/xlist/pkg/builder"
)

var testbuilder1 = []listbuilder.ListDef{
	{
		ID:        "id-list1",
		Class:     "list1",
		Resources: []xlist.Resource{xlist.IPv4},
		Source:    "source list1",
	},
	{
		ID:        "id-list2",
		Class:     "list2",
		Resources: []xlist.Resource{xlist.IPv4},
	},
	{
		ID:        "id-list3",
		Class:     "list2",
		Resources: []xlist.Resource{xlist.IPv4},
	},
}

func TestBuilderBasic(t *testing.T) {
	//register builders
	listbuilder.Register("list1", testBuilderList())

	builder := listbuilder.New()

	//check build registered
	bl, err := builder.Build(testbuilder1[0])
	if err != nil {
		t.Fatalf("building component %v: %v", testbuilder1[0].ID, err)
	}
	resp, err := bl.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != nil {
		t.Errorf("checking list %v: %v", testbuilder1[0].ID, err)
	}
	if !resp.Result {
		t.Errorf("unexpected result in check on list %v: %v", testbuilder1[0].ID, err)
	}
	list1, ok := builder.List(testbuilder1[0].ID)
	if !ok {
		t.Fatalf("returning list %v from builder", testbuilder1[0].ID)
	}
	if bl != list1 {
		t.Fatalf("references mismatch returning list %v", testbuilder1[0].ID)
	}
	_, ok = builder.List("noexists")
	if ok {
		t.Error("returned ok for non existing list")
	}

	//constructing other component
	_, err = builder.Build(testbuilder1[1])
	if err == nil {
		t.Errorf("non error building non existing component")
	}
	//register builder
	listbuilder.Register("list2", testBuilderList())

	bl, err = builder.Build(testbuilder1[1])
	if err != nil {
		t.Fatalf("building component %v: %v", testbuilder1[1].ID, err)
	}
	resp, err = bl.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != nil {
		t.Errorf("checking list %v: %v", testbuilder1[1].ID, err)
	}
	if resp.Result {
		t.Errorf("unexpected result in check on list %v: %v", testbuilder1[1].ID, err)
	}
	list2, ok := builder.List(testbuilder1[1].ID)
	if !ok {
		t.Fatalf("returning list %v from builder", testbuilder1[1].ID)
	}
	if bl != list2 {
		t.Fatalf("references mismatch returning list %v", testbuilder1[1].ID)
	}
}

func TestBuilderStartup(t *testing.T) {
	startValues := make([]bool, 3, 3)
	shutdValues := make([]bool, 3, 3)

	testerr := errors.New("error")
	builder := listbuilder.New()
	builder.OnStartup(func() error { startValues[0] = true; return nil })
	builder.OnShutdown(func() error { shutdValues[0] = true; return nil })
	builder.OnStartup(func() error { return testerr })
	builder.OnShutdown(func() error { return testerr })
	builder.OnStartup(func() error { startValues[2] = true; return nil })
	builder.OnShutdown(func() error { shutdValues[2] = true; return nil })

	err := builder.Start()
	if err != testerr {
		t.Errorf("start(): %v", err)
	}
	if !startValues[0] {
		t.Error("startup func 0 don't started")
	}
	if startValues[2] {
		t.Error("startup func 2 started")
	}

	err = builder.Shutdown()
	if err != testerr {
		t.Errorf("shutdown(): %v", err)
	}
	if !shutdValues[0] {
		t.Error("shutdown func 0 don't shutdown")
	}
	if !shutdValues[2] {
		t.Error("shutdown func 2 don't shudown")
	}
}

var testbuilder2 = []listbuilder.ListDef{
	{
		ID:        "id-list1",
		Class:     "list",
		Resources: []xlist.Resource{xlist.IPv4},
		Source:    "source list1",
	},
	{
		ID:        "id-list2",
		Class:     "comp",
		Resources: []xlist.Resource{xlist.IPv4},
		Contains: []listbuilder.ListDef{
			{
				ID:        "id-list3",
				Class:     "list",
				Resources: []xlist.Resource{xlist.IPv4, xlist.Domain},
			},
			{
				ID:        "id-list4",
				Class:     "list",
				Resources: []xlist.Resource{xlist.IPv4},
			},
			{
				ID:        "id-list5",
				Class:     "comp",
				Resources: []xlist.Resource{xlist.IPv4},
				Contains: []listbuilder.ListDef{
					{
						ID:        "id-list6",
						Class:     "list",
						Resources: []xlist.Resource{xlist.IPv4},
						Source:    "source list6",
					},
					{
						ID:        "id-list7",
						Class:     "list",
						Resources: []xlist.Resource{xlist.IPv4},
						Source:    "source list7",
					},
				},
			},
		},
	},
}

func TestBuilderComp(t *testing.T) {
	//register builders
	listbuilder.Register("list", testBuilderList())
	listbuilder.Register("comp", testBuilderCompo())

	builder := listbuilder.New()
	for _, def := range testbuilder2 {
		_, err := builder.Build(def)
		if err != nil {
			t.Errorf("creating lists: %v", err)
		}
	}
	for i := 1; i <= 7; i++ {
		listID := fmt.Sprintf("id-list%v", i)
		_, ok := builder.List(listID)
		if !ok {
			t.Fatalf("can't get list %s", listID)
		}
	}
	var tests = []struct {
		input string
		want  string
	}{
		{"id-list1", "source list1"},
		{"id-list2", "source list6"},
		{"id-list3", ""},
		{"id-list4", ""},
		{"id-list5", "source list6"},
		{"id-list6", "source list6"},
		{"id-list7", "source list7"},
	}
	for _, test := range tests {
		list, ok := builder.List(test.input)
		if !ok {
			t.Fatalf("can't get list %s", test.input)
		}
		got, err := list.Check(context.Background(), "10.10.10.10", xlist.IPv4)
		if err != nil {
			t.Errorf("error checking list %v: %v", test.input, err)
		}
		if got.Reason != test.want {
			t.Errorf("unexpected check result in list %v: %v", test.input, got.Reason)
		}
	}
}

var testbuilderbad1 = []listbuilder.ListDef{
	{
		ID:        "id-list1",
		Class:     "comp",
		Resources: []xlist.Resource{xlist.IPv4},
		Contains: []listbuilder.ListDef{
			{
				ID:        "id-list2",
				Class:     "list",
				Resources: []xlist.Resource{xlist.IPv4, xlist.Domain},
				Source:    "source list3",
			},
			{
				ID:        "id-list3",
				Class:     "comp",
				Resources: []xlist.Resource{xlist.IPv4},
				Contains: []listbuilder.ListDef{
					{
						ID:        "id-list1",
						Class:     "list",
						Resources: []xlist.Resource{xlist.IPv4},
					},
				},
			},
		},
	},
}

func TestBuilderRecursion(t *testing.T) {
	//register builders
	listbuilder.Register("list", testBuilderList())
	listbuilder.Register("comp", testBuilderCompo())

	builder := listbuilder.New()

	_, err := builder.Build(testbuilderbad1[0])
	if err == nil {
		t.Error("builder fails detecting recursion")
	}
	if !strings.Contains(err.Error(), "loop detection") {
		t.Errorf("error detected diferent from loop detection")
	}
}

var testbuilder3 = []listbuilder.ListDef{
	{
		ID:        "id-list1",
		Class:     "comp",
		Resources: []xlist.Resource{xlist.IPv4},
		Wrappers: []listbuilder.WrapperDef{
			{
				Class: "wrap",
				Opts:  map[string]interface{}{"preffix": "wrapp1-1"},
			},
			{
				Class: "wrap",
				Opts:  map[string]interface{}{"preffix": "wrapp1-2"},
			},
		},
		Contains: []listbuilder.ListDef{
			{
				ID:        "id-list2",
				Class:     "list",
				Resources: []xlist.Resource{xlist.IPv4, xlist.Domain},
			},
			{
				ID:        "id-list3",
				Class:     "list",
				Resources: []xlist.Resource{xlist.IPv4},
				Source:    "source list3",
				Wrappers: []listbuilder.WrapperDef{
					{
						Class: "wrapx",
						Opts:  map[string]interface{}{"preffix": "wrapp3"},
					},
				},
			},
		},
	},
}

func TestBuilderWrapper(t *testing.T) {
	//register builders
	listbuilder.Register("list", testBuilderList())
	listbuilder.Register("comp", testBuilderCompo())
	listbuilder.RegisterWrapper("wrap", testBuilderWrap())

	builder := listbuilder.New()

	_, err := builder.Build(testbuilder3[0])
	if err == nil {
		t.Fatalf("wrapper builds without register: %v", err)
	}

	//register builders
	listbuilder.Register("list", testBuilderList())
	listbuilder.Register("comp", testBuilderCompo())
	listbuilder.RegisterWrapper("wrap", testBuilderWrap())
	listbuilder.RegisterWrapper("wrapx", testBuilderWrap())

	builder = listbuilder.New()
	_, err = builder.Build(testbuilder3[0])
	if err != nil {
		t.Fatalf("unexpected error building list: %v", err)
	}
	for i := 1; i <= 3; i++ {
		listID := fmt.Sprintf("id-list%v", i)
		_, ok := builder.List(listID)
		if !ok {
			t.Fatalf("can't get list %s", listID)
		}
	}

	//make tests
	var tests = []struct {
		input string
		want  string
	}{
		{"id-list1", "wrapp1-2: wrapp1-1: wrapp3: source list3"},
		{"id-list2", ""},
		{"id-list3", "wrapp3: source list3"},
	}
	for _, test := range tests {
		list, ok := builder.List(test.input)
		if !ok {
			t.Fatalf("can't get list %s", test.input)
		}
		got, err := list.Check(context.Background(), "10.10.10.10", xlist.IPv4)
		if err != nil {
			t.Errorf("error checking list %v: %v", test.input, err)
		}
		if got.Reason != test.want {
			t.Errorf("unexpected check result in list %v: %v", test.input, got.Reason)
		}
	}
}
