// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlistd_test

import (
	"sort"
	"testing"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/xlist/pkg/xlistd"
)

func TestCategoryIsValid(t *testing.T) {
	var tests = []struct {
		input int
		want  bool
	}{
		{int(xlistd.Blacklist), true},
		{int(xlistd.Whitelist), true},
		{int(xlistd.Mixedlist), true},
		//invalid values as the time of the writting if the test:
		{-1, false},
		{4, false},
		{10, false},
	}
	for _, test := range tests {
		category := xlistd.Category(test.input)
		if got := category.IsValid(); got != test.want {
			t.Errorf("category[%v].IsValid() = %v", test.input, got)
		}
	}
}

func TestCategoryString(t *testing.T) {
	var tests = []struct {
		input xlistd.Category
		want  string
	}{
		{xlistd.Blacklist, "blacklist"},
		{xlistd.Whitelist, "whitelist"},
		{xlistd.Mixedlist, "mixedlist"},
		{xlistd.Category(-1), "unkown(-1)"},
	}
	for _, test := range tests {
		if got := test.input.String(); got != test.want {
			t.Errorf("category[%v].String() = %v", test.input, got)
		}
	}
}

var testfilters1 = []xlistd.ListDef{
	{
		ID:        "list1",
		Class:     "class1",
		Name:      "Nombre C",
		Category:  xlistd.Blacklist,
		Tags:      []string{"spam", "botnet"},
		Resources: []xlist.Resource{xlist.IPv4},
	},
	{
		ID:        "list2",
		Class:     "class2",
		Name:      "Nombre A",
		Category:  xlistd.Blacklist,
		Tags:      []string{"botnet"},
		Resources: []xlist.Resource{xlist.IPv4},
	},
	{
		ID:        "list3",
		Class:     "class2",
		Name:      "Nombre B",
		Category:  xlistd.Blacklist,
		Tags:      []string{"spam"},
		Resources: []xlist.Resource{xlist.IPv4, xlist.IPv6},
	},
	{
		ID:        "list4",
		Class:     "class2",
		Name:      "Nombre D",
		Category:  xlistd.Whitelist,
		Tags:      []string{},
		Resources: []xlist.Resource{xlist.IPv4},
	},
}

func TestFilterID(t *testing.T) {
	var tests = []struct {
		database []xlistd.ListDef
		listID   string
		want     bool
		wantList xlistd.ListDef
	}{
		{testfilters1, "noexiste", false, xlistd.ListDef{}},
		{testfilters1, "list2", true, testfilters1[1]},
		{testfilters1, "list4", true, testfilters1[3]},
	}
	for _, test := range tests {
		gotList, got := xlistd.FilterID(test.listID, test.database)
		if test.want != got {
			t.Errorf("FilterID(%v, database) = (%v, %v)", test.listID, gotList, got)
		} else {
			if test.wantList.ID != gotList.ID {
				t.Errorf("FilterID(%v, database) = (%v, %v)", test.listID, gotList, got)
			}
		}
	}
}

func TestFilterResource(t *testing.T) {
	var tests = []struct {
		database  []xlistd.ListDef
		resource  xlist.Resource
		want      int
		wantLists []xlistd.ListDef
	}{
		{testfilters1, xlist.Domain, 0, []xlistd.ListDef{}},
		{testfilters1, xlist.IPv6, 1, []xlistd.ListDef{testfilters1[2]}},
		{testfilters1, xlist.IPv4, 4, testfilters1},
	}
	for _, test := range tests {
		gotLists := xlistd.FilterResource(test.resource, test.database)
		if len(gotLists) != test.want {
			t.Errorf("FilterResource(%v, database) = len(%v)", test.resource, len(gotLists))
		} else {
			if !cmpListDefs(gotLists, test.wantLists) {
				t.Errorf("FilterResource(%v, database) = %v", test.resource, gotLists)
			}
		}
	}
}

func TestFilterClass(t *testing.T) {
	var tests = []struct {
		database  []xlistd.ListDef
		class     string
		want      int
		wantLists []xlistd.ListDef
	}{
		{testfilters1, "class69", 0, []xlistd.ListDef{}},
		{testfilters1, "class1", 1, []xlistd.ListDef{testfilters1[0]}},
		{testfilters1, "class2", 3,
			[]xlistd.ListDef{testfilters1[1], testfilters1[2], testfilters1[3]}},
	}
	for _, test := range tests {
		gotLists := xlistd.FilterClass(test.class, test.database)
		if len(gotLists) != test.want {
			t.Errorf("FilterClass(%v, database) = len(%v)", test.class, len(gotLists))
		} else {
			if !cmpListDefs(gotLists, test.wantLists) {
				t.Errorf("FilterClass(%v, database) = %v", test.class, gotLists)
			}
		}
	}
}

func TestFilterCategory(t *testing.T) {
	var tests = []struct {
		database  []xlistd.ListDef
		category  xlistd.Category
		want      int
		wantLists []xlistd.ListDef
	}{
		{testfilters1, xlistd.Mixedlist, 0, []xlistd.ListDef{}},
		{testfilters1, xlistd.Whitelist, 1, []xlistd.ListDef{testfilters1[3]}},
		{testfilters1, xlistd.Blacklist, 3,
			[]xlistd.ListDef{testfilters1[0], testfilters1[1], testfilters1[2]}},
	}
	for _, test := range tests {
		gotLists := xlistd.FilterCategory(test.category, test.database)
		if len(gotLists) != test.want {
			t.Errorf("FilterCategory(%v, database) = len(%v)", test.category, len(gotLists))
		} else {
			if !cmpListDefs(gotLists, test.wantLists) {
				t.Errorf("FilterCategory(%v, database) = %v", test.category, gotLists)
			}
		}
	}
}

func TestFilterTag(t *testing.T) {
	var tests = []struct {
		database  []xlistd.ListDef
		tag       string
		want      int
		wantLists []xlistd.ListDef
	}{
		{testfilters1, "noexiste", 0, []xlistd.ListDef{}},
		{testfilters1, "", 1, []xlistd.ListDef{testfilters1[3]}},
		{testfilters1, "botnet", 2, []xlistd.ListDef{testfilters1[0], testfilters1[1]}},
		{testfilters1, "spam", 2, []xlistd.ListDef{testfilters1[0], testfilters1[2]}},
	}
	for _, test := range tests {
		gotLists := xlistd.FilterTag(test.tag, test.database)
		if len(gotLists) != test.want {
			t.Errorf("FilterTag(%v, database) = len(%v)", test.tag, len(gotLists))
		} else {
			if !cmpListDefs(gotLists, test.wantLists) {
				t.Errorf("FilterTag(%v, database) = %v", test.tag, gotLists)
			}
		}
	}
}

func TestSortedListDefsByID(t *testing.T) {
	var tests = []struct {
		input []xlistd.ListDef
		want  []xlistd.ListDef
	}{
		{testfilters1, testfilters1},
		{[]xlistd.ListDef{
			testfilters1[1], testfilters1[0],
			testfilters1[2], testfilters1[3]}, testfilters1},
	}
	for _, test := range tests {
		gotLists := xlistd.ListDefsByID(test.input)
		sort.Sort(&gotLists)
		if !cmpListDefs(gotLists, test.want) {
			t.Error("ListDefsByID() missmatch")
		}
	}
}

func TestSortedListDefsByName(t *testing.T) {
	var tests = []struct {
		input []xlistd.ListDef
		want  []xlistd.ListDef
	}{
		{testfilters1, []xlistd.ListDef{
			testfilters1[1], testfilters1[2],
			testfilters1[0], testfilters1[3]}},

		{[]xlistd.ListDef{
			testfilters1[3], testfilters1[2],
			testfilters1[1], testfilters1[0]},
			[]xlistd.ListDef{
				testfilters1[1], testfilters1[2],
				testfilters1[0], testfilters1[3]}},
	}
	for _, test := range tests {
		gotLists := xlistd.ListDefsByName(test.input)
		sort.Sort(&gotLists)
		if !cmpListDefs(gotLists, test.want) {
			t.Error("ListDefsByName() missmatch")
		}
	}
}

//TODO check ListDefsFromFile

func cmpListDefs(a, b []xlistd.ListDef) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v.ID != b[i].ID {
			return false
		}
	}
	return true
}
