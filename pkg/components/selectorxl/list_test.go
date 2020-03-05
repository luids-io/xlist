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
	selector := selectorxl.New()
	resp, err := selector.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != xlist.ErrNotImplemented {
		t.Fatal("selector.Check unexpected error")
	}
	rblip := &mockxl.List{
		ResourceList: []xlist.Resource{xlist.IPv4, xlist.IPv6},
		Results:      []bool{true}, Reason: "ip",
	}
	selector.SetService(xlist.IPv6, rblip)
	resp, err = selector.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != xlist.ErrNotImplemented {
		t.Fatal("selector.Check unexpected error")
	}

	selector.SetService(xlist.IPv4, rblip)
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
		ResourceList: []xlist.Resource{xlist.Domain},
		Results:      []bool{true}, Reason: "domain",
	}
	selector.SetService(xlist.Domain, rbldomain)
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
	type item struct {
		resource xlist.Resource
		list     xlist.List
	}
	var tests = []struct {
		checkers []item
		wantErr  bool
	}{
		{[]item{}, false}, //0
		{[]item{
			{resource: xlist.IPv4, list: rblOk},
		}, false}, //1
		{[]item{
			{resource: xlist.IPv4, list: rblOk},
			{resource: xlist.Domain, list: rblOk},
		}, false}, //2
		{[]item{
			{resource: xlist.IPv4, list: rblFail},
		}, true}, //3
		{[]item{
			{resource: xlist.IPv4, list: rblOk},
			{resource: xlist.Domain, list: rblFail},
		}, true}, //4
		{[]item{
			{resource: xlist.IPv4, list: rblFail},
			{resource: xlist.Domain, list: rblOk},
		}, true}, //5
	}
	for idx, test := range tests {
		slist := selectorxl.New()
		for _, c := range test.checkers {
			slist.SetService(c.resource, c.list)
		}
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
		checkers []item
		want     []xlist.Resource
	}{
		{[]item{}, []xlist.Resource{}}, //0
		{[]item{
			{resource: xlist.IPv4, list: rblOk},
		}, onlyIPv4}, //1
		{[]item{
			{resource: xlist.IPv4, list: rblOk},
			{resource: xlist.Domain, list: rblOk},
		}, []xlist.Resource{xlist.IPv4, xlist.Domain}}, //2
		{[]item{
			{resource: xlist.IPv4, list: rblOk},
			{resource: xlist.IPv4, list: rblOk},
		}, onlyIPv4}, //3
		{[]item{
			{resource: xlist.IPv6, list: rblOk},
			{resource: xlist.IPv4, list: rblOk},
		}, onlyIP}, //4
	}
	for idx, test := range tests {
		slist := selectorxl.New()
		for _, c := range test.checkers {
			slist.SetService(c.resource, c.list)
		}
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
	ip4 := &mockxl.List{
		Results: []bool{true}, ResourceList: []xlist.Resource{xlist.IPv4},
		Reason: "ip4"}
	domain := &mockxl.List{
		Results: []bool{true}, ResourceList: []xlist.Resource{xlist.Domain},
		Reason: "domain"}

	rbl := selectorxl.New()
	rbl.SetService(xlist.IPv4, ip4)
	rbl.SetService(xlist.Domain, domain)

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
