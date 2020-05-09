// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package wbeforexl_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/components/mockxl"
	"github.com/luids-io/xlist/pkg/components/wbeforexl"
)

func TestList_Check(t *testing.T) {
	rblFalse := &mockxl.List{ResourceList: onlyIPv4}
	rblTrue := &mockxl.List{ResourceList: onlyIPv4, Results: []bool{true}}
	rblFail := &mockxl.List{ResourceList: onlyIPv4, Fail: true}

	var tests = []struct {
		resources []xlist.Resource
		white     xlist.List
		black     xlist.List
		want      bool
		wantErr   bool
	}{
		{onlyIPv4, rblFalse, rblFalse, false, false},
		{onlyIPv4, rblTrue, rblFalse, false, false},
		{onlyIPv4, rblFalse, rblTrue, true, false},
		{onlyIPv4, rblTrue, rblTrue, false, false},
		// errors
		{onlyDomain, rblFalse, rblFalse, false, true},
		{onlyIPv4, rblFail, rblFalse, false, true},
		{onlyIPv4, rblTrue, rblFail, false, false},
		{onlyIPv4, rblFalse, rblFail, false, true},
	}
	for idx, test := range tests {
		wblist := wbeforexl.New("test", test.white, test.black, test.resources, wbeforexl.Config{})
		resp, err := wblist.Check(context.Background(), "10.10.10.10", xlist.IPv4)
		if test.wantErr && err == nil {
			t.Errorf("wbefore.Check idx[%v] expected error", idx)
		} else if !test.wantErr && err != nil {
			t.Errorf("wbefore.Check idx[%v] unexpected error: %v", idx, err)
		}
		if test.want != resp.Result {
			t.Errorf("wbefore.Check idx[%v] want=%v got=%v", idx, test.want, resp.Result)
		}
	}
}

func TestList_CheckCancel(t *testing.T) {
	white := &mockxl.List{
		ResourceList: onlyIPv4,
		Sleep:        100 * time.Millisecond,
	}
	black := &mockxl.List{
		ResourceList: onlyIPv4,
		Results:      []bool{true},
	}

	wblist := wbeforexl.New("test", white, black, onlyIPv4, wbeforexl.Config{})
	resp, err := wblist.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != nil {
		t.Errorf("wbefore.Check unexpected error: %v", err)
	}
	if !resp.Result {
		t.Errorf("wbefore.Check want=%v got=%v", true, resp.Result)
	}

	ctxtimeout, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	resp, err = wblist.Check(ctxtimeout, "10.10.10.10", xlist.IPv4)
	if err == nil {
		t.Error("wbefore.Check expected error")
	}
	if err != xlist.ErrCanceledRequest {
		t.Errorf("wbefore.Check unexpected error: %v", err)
	}
}

func TestList_Ping(t *testing.T) {
	rblOk := &mockxl.List{ResourceList: onlyIPv4, Fail: false}
	rblFail := &mockxl.List{ResourceList: onlyIPv4, Fail: true}

	var tests = []struct {
		white   xlist.List
		black   xlist.List
		wantErr bool
	}{
		{rblOk, rblOk, false},    //0
		{rblOk, rblFail, true},   //1
		{rblFail, rblOk, true},   //2
		{rblFail, rblFail, true}, //3
	}
	for idx, test := range tests {
		wblist := wbeforexl.New("test", test.white, test.black, onlyIPv4, wbeforexl.Config{})
		err := wblist.Ping()
		switch {
		case test.wantErr && err == nil:
			t.Errorf("wbefore.Ping idx[%v] expected error", idx)
		case !test.wantErr && err != nil:
			t.Errorf("wbefore.Ping idx[%v] unexpected error: %v", idx, err)
		}
	}
}

func TestList_Resources(t *testing.T) {
	var tests = []struct {
		in   []xlist.Resource
		want []xlist.Resource
	}{
		{[]xlist.Resource{}, []xlist.Resource{}},
		{[]xlist.Resource{xlist.Domain}, onlyDomain},
		{[]xlist.Resource{xlist.IPv4, xlist.IPv4}, onlyIPv4},
		{[]xlist.Resource{xlist.IPv4, xlist.IPv6, xlist.IPv6}, onlyIP},
		{xlist.Resources, xlist.Resources},
	}
	for idx, test := range tests {
		wblist := wbeforexl.New("test", &mockxl.List{}, &mockxl.List{}, test.in, wbeforexl.Config{})
		got := wblist.Resources()
		if !cmpResourceSlice(got, test.want) {
			t.Errorf("idx[%v] wbefore.Resources() got=%v want=%v", idx, got, test.want)
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
	resources := []xlist.Resource{xlist.IPv4}
	white := &mockxl.List{
		Identifier:   "whitelist",
		ResourceList: resources,
		Results:      []bool{true, false},
	}
	black := &mockxl.List{
		Identifier:   "blacklist",
		ResourceList: resources,
		Results:      []bool{true},
	}

	//constructs wbefore rbl
	rbl := wbeforexl.New("test", white, black, resources, wbeforexl.Config{Reason: "hello"})
	//in the first check, whitelist returns true -> return false
	resp, err := rbl.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != nil && resp.Result {
		log.Fatalln("this should not happen")
	}
	fmt.Println("check 1:", resp.Result)

	//in the second check, whitelist returns false -> blacklist is checked
	// and blacklist allways returns true -> return true
	resp, err = rbl.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != nil && !resp.Result {
		log.Fatalln("this should not happen")
	}
	fmt.Println("check 2:", resp.Result, resp.Reason)

	// Output:
	//check 1: false
	//check 2: true hello
}
