// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package cachewr_test

import (
	"context"
	"testing"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/components/mockxl"
	"github.com/luids-io/xlist/pkg/wrappers/cachewr"
)

func TestWrapper_Check(t *testing.T) {
	ip4 := []xlist.Resource{xlist.IPv4}
	mockup := &mockxl.List{ResourceList: ip4, Results: []bool{true, false}}
	cache := cachewr.New(mockup, cachewr.DefaultConfig())

	var tests = []struct {
		name      string
		want      bool
		wantCache bool
	}{
		{"10.10.10.1", true, false},  //0
		{"10.10.10.2", false, false}, //1
		{"10.10.10.3", true, false},  //2
		{"10.10.10.4", false, false}, //3
		//test cache
		{"10.10.10.1", true, true},  //4
		{"10.10.10.3", true, true},  //5
		{"10.10.10.2", false, true}, //6
		{"10.10.10.4", false, true}, //7
		{"10.10.10.1", true, true},  //8
		//
		{"192.168.10.1", true, false}, //9
		{"192.168.10.1", true, true},  //10
	}
	for idx, test := range tests {
		resp, err := cache.Check(context.Background(), test.name, xlist.IPv4)
		if err != nil {
			t.Errorf("idx[%v] cache.Check(): err=%v", idx, err)
		}
		if test.want != resp.Result {
			t.Errorf("idx[%v] cache.Check(): want=%v got=%v", idx, test.want, resp)
		}
		if test.wantCache && resp.TTL <= 0 {
			t.Errorf("idx[%v] cache.Check(): wantCache=%v got=%v", idx, test.wantCache, resp)
		}
	}
}

func TestWrapper_CheckNegative(t *testing.T) {
	ip4 := []xlist.Resource{xlist.IPv4}
	mockup := &mockxl.List{ResourceList: ip4, Results: []bool{true, false}}
	cfg := cachewr.DefaultConfig()
	cfg.NegativeTTL = xlist.NeverCache
	cache := cachewr.New(mockup, cfg)

	var tests = []struct {
		name      string
		want      bool
		wantCache bool
	}{
		{"10.10.10.1", true, false},  //0 //flag: false
		{"10.10.10.2", false, false}, //1 //flag: true
		{"10.10.10.3", true, false},  //2 //flag: false
		{"10.10.10.4", false, false}, //3 //flag: true

		//test cache
		{"10.10.10.1", true, true},   //4
		{"10.10.10.3", true, true},   //5
		{"10.10.10.2", true, false},  //6 //flag: false
		{"10.10.10.4", false, false}, //7 //flag: true
		{"10.10.10.1", true, true},   //8
		//
		{"192.168.10.1", true, false}, //9 //flag: false
		{"192.168.10.1", true, true},  //10
	}
	for idx, test := range tests {
		resp, err := cache.Check(context.Background(), test.name, xlist.IPv4)
		if err != nil {
			t.Errorf("idx[%v] cache.Check(): err=%v", idx, err)
		}
		if test.want != resp.Result {
			t.Errorf("idx[%v] cache.Check(): want=%v got=%v", idx, test.want, resp)
		}
		if test.wantCache && resp.TTL <= 0 {
			t.Errorf("idx[%v] cache.Check(): wantCache=%v got=%v", idx, test.wantCache, resp)
		}
	}
}

//TODO: checks for cleanups
