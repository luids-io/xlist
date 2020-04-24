// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package memxl_test

import (
	"context"
	"testing"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/components/memxl"
)

func TestList_Check_domains(t *testing.T) {
	list := memxl.New([]xlist.Resource{xlist.Domain}, memxl.Config{})
	err := list.AddDomains([]string{"www.micasa.com", "www.tucasa.com"})
	if err != nil {
		t.Fatalf("memxl.AddDomains(): err=%v", err)
	}
	var tests = []struct {
		name string
		want bool
	}{
		{"barrapunto.com", false},      //0
		{"www.micasa.com", true},       //1
		{"kaka.es", false},             //2
		{"algo.www.micasa.com", false}, //3
		{"www.tucasa.com", true},       //4
	}
	//run tests
	for idx, test := range tests {
		got, err := list.Check(context.Background(), test.name, xlist.Domain)
		if err != nil {
			t.Fatalf("unexpected error")
		}
		if got.Result != test.want {
			t.Errorf("idx[%v] memxl.Check(): want=%v got=%v", idx, test.want, got.Result)
		}
	}
}

func TestList_Check_subdomains1(t *testing.T) {
	list := memxl.New([]xlist.Resource{xlist.Domain}, memxl.Config{})
	err := list.AddDomains([]string{"www.malware.com"})
	if err != nil {
		t.Fatalf("memxl.AddDomains(): err=%v", err)
	}
	err = list.AddSubdomains([]string{"es", "org"})
	if err != nil {
		t.Fatalf("memxl.AddSubdomains(): err=%v", err)
	}
	var tests = []struct {
		name string
		want bool
	}{
		{"barrapunto.com", false},                  //0
		{"www.micasa.com", false},                  //1
		{"kaka.es", true},                          //2
		{"algo.www.micasa.com", false},             //3
		{"malware.com", false},                     //4
		{"www.malware.org", true},                  //5
		{"www.malware.com", true},                  //6
		{"esto.no.debe.de.estar.porque.eo", false}, //7
		{"com", false},                             //8
	}
	//run tests
	for idx, test := range tests {
		got, err := list.Check(context.Background(), test.name, xlist.Domain)
		if err != nil {
			t.Fatalf("unexpected error")
		}
		if got.Result != test.want {
			t.Errorf("idx[%v] memxl.Check(): want=%v got=%v", idx, test.want, got.Result)
		}
	}
}

func TestList_Check_subdomains2(t *testing.T) {
	list := memxl.New([]xlist.Resource{xlist.Domain}, memxl.Config{})
	err := list.AddDomains([]string{"www.micasa.com"})
	if err != nil {
		t.Fatalf("memxl.AddDomains(): err=%v", err)
	}
	err = list.AddSubdomains([]string{"es", "malware.com", "www.tucasa.com"})
	if err != nil {
		t.Fatalf("memxl.AddSubdomains(): err=%v", err)
	}
	err = list.AddDomains([]string{"www.tucasa.com"})
	if err != nil {
		t.Fatalf("memxl.AddDomains(): err=%v", err)
	}

	var tests = []struct {
		name string
		want bool
	}{
		{"barrapunto.com", false},                    //0
		{"www.micasa.com", true},                     //1
		{"kaka.es", true},                            //2
		{"algo.www.micasa.com", false},               //3
		{"www.tucasa.com", true},                     //4
		{"algo.www.tucasa.com", true},                //5
		{"mas.algo.www.tucasa.com", true},            //6
		{"pero.mucho.mas.algo.www.tucasa.com", true}, //7
		{"esto.no.debe.de.estar.porque.eo", false},   //8
	}
	//run tests
	for idx, test := range tests {
		got, err := list.Check(context.Background(), test.name, xlist.Domain)
		if err != nil {
			t.Fatalf("unexpected error")
		}
		if got.Result != test.want {
			t.Errorf("idx[%v] memxl.Check(): want=%v got=%v", idx, test.want, got.Result)
		}
	}
}
