// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package responsewr_test

import (
	"context"
	"testing"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/components/mockxl"
	"github.com/luids-io/xlist/pkg/wrappers/responsewr"
)

func TestWrapper_CheckNegate(t *testing.T) {
	ip4 := []xlist.Resource{xlist.IPv4}
	mockup := &mockxl.List{ResourceList: ip4, Results: []bool{true, false}}
	respwr := responsewr.New(mockup, responsewr.Negate(true))

	var tests = []struct {
		name    string
		want    bool
		wantErr bool
	}{
		{"10.10.10.1", false, false},    //0
		{"10.10.10.2", true, false},     //1
		{"10.10.10.3", false, false},    //2
		{"10.10.10.4", true, false},     //3
		{"10.10.10.5", false, false},    //4
		{"www.google.com", false, true}, //5
	}
	for idx, test := range tests {
		resp, err := respwr.Check(context.Background(), test.name, xlist.IPv4)
		if err != nil && !test.wantErr {
			t.Errorf("idx[%v] respwr.Check(): err=%v", idx, err)
		}
		if test.want != resp.Result {
			t.Errorf("idx[%v] respwr.Check(): want=%v got=%v", idx, test.want, resp)
		}
	}
}

func TestWrapper_CheckTTL(t *testing.T) {
	ip4 := []xlist.Resource{xlist.IPv4}
	mockup := &mockxl.List{ResourceList: ip4, Results: []bool{true, false}, TTL: 600}

	var tests = []struct {
		ttl  int
		want int
	}{
		{0, 600},                             //0
		{10, 10},                             //1
		{xlist.NeverCache, xlist.NeverCache}, //2
		{-2, 600},                            //3
	}
	for idx, test := range tests {
		respwr := responsewr.New(mockup, responsewr.TTL(test.ttl))
		resp, _ := respwr.Check(context.Background(), "10.10.10.10", xlist.IPv4)
		if test.want != resp.TTL {
			t.Errorf("idx[%v] respwr.Check(): want=%v got=%v", idx, test.want, resp.TTL)
		}
	}
}

func TestWrapper_CheckReason(t *testing.T) {
	ip4 := []xlist.Resource{xlist.IPv4}
	mockup := &mockxl.List{ResourceList: ip4, Results: []bool{true, false}, Reason: "mockup"}
	respwr := responsewr.New(mockup, responsewr.Reason("cambiada"))

	var tests = []struct {
		name string
		want string
	}{
		{"10.10.10.1", "cambiada"}, //0
		{"10.10.10.2", ""},         //1
		{"10.10.10.3", "cambiada"}, //0
	}
	for idx, test := range tests {
		resp, _ := respwr.Check(context.Background(), test.name, xlist.IPv4)
		if test.want != resp.Reason {
			t.Errorf("idx[%v] respwr.Check(): want=%v got=%v", idx, test.want, resp.Reason)
		}
	}
}

func TestWrapper_CheckPrefix(t *testing.T) {
	ip4 := []xlist.Resource{xlist.IPv4}
	mockup := &mockxl.List{ResourceList: ip4, Results: []bool{true, false}, Reason: "mockup"}
	respwr := responsewr.New(mockup, responsewr.PreffixReason("prueba"))

	var tests = []struct {
		name string
		want string
	}{
		{"10.10.10.1", "prueba: mockup"}, //0
		{"10.10.10.2", ""},               //1
		{"10.10.10.3", "prueba: mockup"}, //0
	}
	for idx, test := range tests {
		resp, _ := respwr.Check(context.Background(), test.name, xlist.IPv4)
		if test.want != resp.Reason {
			t.Errorf("idx[%v] respwr.Check(): want=%v got=%v", idx, test.want, resp.Reason)
		}
	}
}
