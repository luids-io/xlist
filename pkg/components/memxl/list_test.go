// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package memxl_test

import (
	"context"
	"strings"
	"testing"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/components/memxl"
)

func TestList_Check(t *testing.T) {
	list := getFilledList(t)

	var tests = []struct {
		name     string
		resource xlist.Resource
		want     bool
		wantErr  bool
	}{
		{"8.8.8.8", xlist.IPv4, false, false},             //0
		{"1.1.1.1", xlist.IPv4, true, false},              //1
		{"127.1.1.1", xlist.IPv4, true, false},            //2
		{"barrapunto.com", xlist.Domain, false, false},    //3
		{"www.micasa.com", xlist.Domain, true, false},     //4
		{"kaka.es", xlist.Domain, true, false},            //5
		{"algo.de.sucasa.com", xlist.Domain, true, false}, //6
		//errs
		{"www.google.com", xlist.IPv4, false, true},            //7
		{"fe80::3289:ad8e:8259:c878", xlist.IPv6, false, true}, //8
	}
	//run tests
	for idx, test := range tests {
		got, err := list.Check(context.Background(), test.name, test.resource)
		switch {
		case err != nil && !test.wantErr:
			t.Errorf("idx[%v] memxl.Check(): err=%v", idx, err)
		case err == nil && test.wantErr:
			t.Errorf("idx[%v] memxl.Check(): expected error", idx)
		case got.Result != test.want:
			t.Errorf("idx[%v] memxl.Check(): want=%v got=%v", idx, test.want, got)
		}
	}
}

func TestList_Clear(t *testing.T) {
	list := getFilledList(t)
	list.Clear(context.Background())

	var tests = []struct {
		name     string
		resource xlist.Resource
	}{
		{"8.8.8.8", xlist.IPv4},
		{"1.1.1.1", xlist.IPv4},
		{"127.1.1.1", xlist.IPv4},
		{"barrapunto.com", xlist.Domain},
		{"www.micasa.com", xlist.Domain},
		{"kaka.es", xlist.Domain},
		{"algo.de.sucasa.com", xlist.Domain},
	}
	for idx, test := range tests {
		got, err := list.Check(context.Background(), test.name, test.resource)
		if err != nil {
			t.Fatalf("idx[%v] memxl.Check(): err=%v", idx, err)
		}
		if got.Result {
			t.Errorf("idx[%v] memxl.Check(): got=%v", idx, got)
		}
	}
}

var testfile1 = "../../../test/testdata/testfile1.xlist"
var testfileerr = "../../../test/testdata/testfile-err.xlist"

func TestLoadFile(t *testing.T) {
	list := memxl.New("test1",
		[]xlist.Resource{xlist.IPv4, xlist.Domain, xlist.IPv6},
		memxl.Config{})
	list.AddIP4s([]string{"2.2.2.2"})
	err := memxl.LoadFromFile(list, testfile1, true) //with clear
	if err != nil {
		t.Fatalf("memxl.LoadFromFile(): err=%v", err)
	}
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
		{"2001:d00::c878", xlist.IPv6, true},
		{"2000:d00::c878", xlist.IPv6, false},
	}
	//run tests
	for idx, test := range tests {
		got, err := list.Check(context.Background(), test.name, test.resource)
		if err != nil {
			t.Errorf("idx[%v] memxl.Check(): err=%v", idx, err)
		}
		if got.Result != test.want {
			t.Errorf("idx[%v] memxl.Check(): want=%v got=%v", idx, test.want, got)
		}
	}
}

func TestLoadData(t *testing.T) {
	list := memxl.New("test1", []xlist.Resource{xlist.IPv4, xlist.Domain, xlist.IPv6}, memxl.Config{})
	list.AddIP4s([]string{"2.2.2.2"})
	data := []memxl.Data{
		{Resource: xlist.IPv4, Format: xlist.Plain, Value: "11.22.33.44"},
		{Resource: xlist.IPv4, Format: xlist.CIDR, Value: "10.5.0.0/16"},
		{Resource: xlist.IPv6, Format: xlist.CIDR, Value: "2001:d00::/24"},
		{Resource: xlist.Domain, Format: xlist.Sub, Value: "sucasa.com"},
		{Resource: xlist.IPv4, Format: xlist.Plain, Value: "54.37.157.73"},
		{Resource: xlist.IPv4, Format: xlist.Plain, Value: "192.168.10.10"},
		{Resource: xlist.Domain, Format: xlist.Plain, Value: "www.micasa.com"},
		{Resource: xlist.IPv4, Format: xlist.Plain, Value: "1.2.3.4"},
		{Resource: xlist.IPv6, Format: xlist.Plain, Value: "fe80::3289:ad8e:8259:c878"},
	}
	err := memxl.LoadFromData(list, data, true) //with clear
	if err != nil {
		t.Fatalf("memxl.LoadFromFile(): err=%v", err)
	}
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
		{"2001:d00::c878", xlist.IPv6, true},
		{"2000:d00::c878", xlist.IPv6, false},
	}
	//run tests
	for idx, test := range tests {
		got, err := list.Check(context.Background(), test.name, test.resource)
		if err != nil {
			t.Errorf("idx[%v] memxl.Check(): err=%v", idx, err)
		}
		if got.Result != test.want {
			t.Errorf("idx[%v] memxl.Check(): want=%v got=%v", idx, test.want, got)
		}
	}
}

func TestLoadFileErrors(t *testing.T) {
	list := memxl.New("test1", []xlist.Resource{xlist.IPv4, xlist.Domain}, memxl.Config{})
	err := memxl.LoadFromFile(list, "noexistefichero.list", false)

	if err == nil || !strings.Contains(err.Error(), "opening file") {
		t.Fatalf("memxl.LoadFromFile(): expected error %v", err)
	}
	err = memxl.LoadFromFile(list, testfileerr, false)
	if err == nil || !strings.Contains(err.Error(), "ip4,cidr,-10.5.0.0/16") {
		t.Fatalf("memxl.LoadFromFile(): expected error %v", err)
	}
}

func getFilledList(t *testing.T) *memxl.List {
	list := memxl.New("test1", []xlist.Resource{xlist.IPv4, xlist.Domain}, memxl.Config{})
	err := list.AddIP4s([]string{"10.54.1.1", "1.1.1.1", "192.168.1.3"})
	if err != nil {
		t.Fatalf("memxl.AddIP4s(): err=%v", err)
	}
	err = list.AddCIDR4s([]string{"192.168.0.0/24", "127.0.0.0/8"})
	if err != nil {
		t.Fatalf("memxl.AddCIDRs(): err=%v", err)
	}
	err = list.AddDomains([]string{"www.micasa.com", "www.tucasa.com"})
	if err != nil {
		t.Fatalf("memxl.AddDomains(): err=%v", err)
	}
	err = list.AddSubdomains([]string{"es", "sucasa.com"})
	if err != nil {
		t.Fatalf("memxl.AddSubdomains(): err=%v", err)
	}
	return list
}
