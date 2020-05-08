// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package selectorxl_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/components/mockxl"
	"github.com/luids-io/xlist/pkg/components/selectorxl"
)

func TestList_Check(t *testing.T) {
	//create services map
	services := make(map[xlist.Resource]xlist.List)

	selector := selectorxl.New("test", services, selectorxl.Config{})
	resp, err := selector.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != xlist.ErrNotSupported {
		t.Fatal("selector.Check unexpected error")
	}
	rblip := &mockxl.List{
		Identifier:   "ipservice",
		ResourceList: []xlist.Resource{xlist.IPv4, xlist.IPv6},
		Results:      []bool{true}, Reason: "ip",
	}

	services[xlist.IPv6] = rblip
	selector = selectorxl.New("test", services, selectorxl.Config{})
	resp, err = selector.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != xlist.ErrNotSupported {
		t.Fatal("selector.Check unexpected error")
	}

	services[xlist.IPv4] = rblip
	selector = selectorxl.New("test", services, selectorxl.Config{})
	resp, err = selector.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != nil {
		t.Fatalf("selector.Check unexpected error: %v", err)
	}
	if resp.Reason != "ip" {
		t.Errorf("selector.Check ip4 unexpected response: %v", resp)
	}

	resources := selector.Resources()
	if len(resources) != 2 {
		t.Errorf("unexpected resources: %v", resources)
	}
	if !xlist.IPv4.InArray(resources) {
		t.Error("ip4 not in resources")
	}
	if !xlist.IPv6.InArray(resources) {
		t.Error("ip4 not in resources")
	}
	if xlist.Domain.InArray(resources) {
		t.Error("domain in resources")
	}

	rbldomain := &mockxl.List{
		Identifier:   "domainservice",
		ResourceList: []xlist.Resource{xlist.Domain},
		Results:      []bool{true}, Reason: "domain",
	}

	services[xlist.Domain] = rbldomain
	selector = selectorxl.New("test", services, selectorxl.Config{})
	resp, err = selector.Check(context.Background(), "www.google.com", xlist.Domain)
	if err != nil {
		t.Fatalf("selector.Check unexpected error: %v", err)
	}
	if resp.Reason != "domain" {
		t.Errorf("selector.Check domain unexpected response: %v", resp)
	}
}

func TestList_Ping(t *testing.T) {
	rblOk := &mockxl.List{ResourceList: xlist.Resources, Fail: false}
	rblFail := &mockxl.List{ResourceList: xlist.Resources, Fail: true}
	var tests = []struct {
		checkers map[xlist.Resource]xlist.List
		wantErr  bool
	}{
		{map[xlist.Resource]xlist.List{}, false}, //0
		{map[xlist.Resource]xlist.List{
			xlist.IPv4: rblOk,
		}, false}, //1
		{map[xlist.Resource]xlist.List{
			xlist.IPv4:   rblOk,
			xlist.Domain: rblOk,
		}, false}, //2
		{map[xlist.Resource]xlist.List{
			xlist.IPv4: rblFail,
		}, true}, //3
		{map[xlist.Resource]xlist.List{
			xlist.IPv4:   rblOk,
			xlist.Domain: rblFail,
		}, true}, //4
		{map[xlist.Resource]xlist.List{
			xlist.IPv4:   rblFail,
			xlist.Domain: rblOk,
		}, true}, //5
	}
	for idx, test := range tests {
		slist := selectorxl.New("test", test.checkers, selectorxl.Config{})
		err := slist.Ping()
		switch {
		case test.wantErr && err == nil:
			t.Errorf("idx[%v] selector.Ping expected error", idx)
		case !test.wantErr && err != nil:
			t.Errorf("idx[%v] selector.Ping unexpected error: %v", idx, err)
		}
	}
}

func TestList_Resources(t *testing.T) {
	rblOk := &mockxl.List{ResourceList: xlist.Resources}
	type item struct {
		resource xlist.Resource
		list     xlist.List
	}
	var tests = []struct {
		checkers map[xlist.Resource]xlist.List
		want     []xlist.Resource
	}{
		{map[xlist.Resource]xlist.List{}, []xlist.Resource{}}, //0
		{map[xlist.Resource]xlist.List{
			xlist.IPv4: rblOk,
		}, onlyIPv4}, //1
		{map[xlist.Resource]xlist.List{
			xlist.IPv4:   rblOk,
			xlist.Domain: rblOk,
		}, []xlist.Resource{xlist.IPv4, xlist.Domain}}, //2
		{map[xlist.Resource]xlist.List{
			xlist.IPv4: rblOk,
		}, onlyIPv4}, //3
		{map[xlist.Resource]xlist.List{
			xlist.IPv6: rblOk,
			xlist.IPv4: rblOk,
		}, onlyIP}, //4
	}
	for idx, test := range tests {
		slist := selectorxl.New("test", test.checkers, selectorxl.Config{})
		got := slist.Resources()
		if !cmpResourceSlice(got, test.want) {
			t.Errorf("idx[%v] selector.Resources() got=%v want=%v", idx, got, test.want)
		}
	}
}

func cmpResourceSlice(a, b []xlist.Resource) bool {
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

func ExampleList() {
	services := map[xlist.Resource]xlist.List{
		xlist.IPv4: &mockxl.List{
			Identifier: "ipservice",
			Results:    []bool{true}, ResourceList: []xlist.Resource{xlist.IPv4},
			Reason: "ip4",
		},
		xlist.Domain: &mockxl.List{
			Identifier: "domainservice",
			Results:    []bool{true}, ResourceList: []xlist.Resource{xlist.Domain},
			Reason: "domain",
		}}

	rbl := selectorxl.New("test", services, selectorxl.Config{})
	resp, err := rbl.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != nil || !resp.Result {
		log.Fatalln("this should not happen")
	}
	fmt.Println("check 1:", resp.Reason)

	resp, err = rbl.Check(context.Background(), "www.google.com", xlist.Domain)
	if err != nil || !resp.Result {
		log.Fatalln("this should not happen")
	}
	fmt.Println("check 2:", resp.Reason)

	// Output:
	// check 1: ip4
	// check 2: domain
}
