// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package filexl_test

import (
	"context"
	"strings"
	"testing"

	"github.com/luids-io/core/yalogi"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/xlist/pkg/xlistd/components/filexl"
)

var testdir = "../../../../test/testdata"
var testfile1 = "../../../../test/testdata/testfile1.xlist"
var testfileerr = "../../../../test/testdata/testfile-err.xlist"

func TestList_Check(t *testing.T) {
	list := filexl.New("test1", testfile1,
		[]xlist.Resource{xlist.IPv4, xlist.Domain, xlist.IPv6},
		filexl.Config{}, yalogi.LogNull,
	)
	err := list.Open()
	if err != nil {
		t.Fatalf("filexl.Start(): err=%v", err)
	}
	defer list.Close()

	var tests = []struct {
		name     string
		resource xlist.Resource
		want     bool
	}{
		{"8.8.8.8", xlist.IPv4, false},
		{"1.1.1.1", xlist.IPv4, false},
		{"2.2.2.2", xlist.IPv4, false},
		{"1.2.3.4", xlist.IPv4, true},
		{"10.5.1.1", xlist.IPv4, true},
		{"barrapunto.com", xlist.Domain, false},
		{"www.micasa.com", xlist.Domain, true},
		{"algo.de.sucasa.com", xlist.Domain, true},
		{"fe80::3289:ad8e:8259:c877", xlist.IPv6, false},
		{"fe80::3289:ad8e:8259:c878", xlist.IPv6, true},
	}
	//run tests
	for idx, test := range tests {
		got, err := list.Check(context.Background(), test.name, test.resource)
		if err != nil {
			t.Errorf("idx[%v] filexl.Check(): err=%v", idx, err)
		}
		if got.Result != test.want {
			t.Errorf("idx[%v] filexl.Check(): want=%v got=%v", idx, test.want, got)
		}
	}
}

func TestList_New(t *testing.T) {
	list := filexl.New("test1", testfileerr,
		[]xlist.Resource{xlist.IPv4, xlist.Domain, xlist.IPv6},
		filexl.Config{}, yalogi.LogNull,
	)
	err := list.Open()
	if err == nil || !strings.Contains(err.Error(), "ip4,cidr,-10.5.0.0/16") {
		t.Fatalf("filexl.Start(): expected error %v", err)
	}

	list = filexl.New("test2", testfile1,
		[]xlist.Resource{xlist.IPv4, xlist.Domain, xlist.IPv6},
		filexl.Config{}, yalogi.LogNull,
	)
	err = list.Ping()
	if err == nil {
		t.Error("filexl.Ping(): expected error")
	}
	err = list.Open()
	if err != nil {
		t.Fatalf("filexl.Start(): err=%v", err)
	}
	defer list.Close()
}

//TODO: check autoreload
