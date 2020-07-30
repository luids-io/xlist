// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package policywr_test

import (
	"context"
	"testing"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/reason"
	"github.com/luids-io/xlist/pkg/xlistd/components/mockxl"
	"github.com/luids-io/xlist/pkg/xlistd/wrappers/policywr"
)

func TestWrapper_Policy(t *testing.T) {
	ip4 := []xlist.Resource{xlist.IPv4}
	mockup := &mockxl.List{ResourceList: ip4, Results: []bool{true}}

	var tests = []struct {
		reason string
		policy string
		want   string
	}{
		{"razon", "", "razon"},                                                                   //0
		{"razon", "[policy][/policy]", "razon"},                                                  //1
		{"razon", "[policy]kk=12[/policy]", "[policy]kk=12[/policy]razon"},                       //2
		{"ra[policy]kk=11[/policy]zon", "", "razon"},                                             //3
		{"razon[policy]kk=11[/policy]", "[policy]kk=12[/policy]", "[policy]kk=12[/policy]razon"}, //4
	}
	for idx, test := range tests {
		mockup.Reason = test.reason
		policy := reason.NewPolicy()
		policy.FromString(test.policy)
		checker := policywr.New(mockup, policy, policywr.Config{})

		got, _ := checker.Check(context.Background(), "10.10.10.10", xlist.IPv4)
		if got.Reason != test.want {
			t.Errorf("idx[%v] policwr.Check(): want=%v got=%v", idx, test.want, got.Reason)
		}
	}
}

func TestWrapper_PolicyMerge(t *testing.T) {
	ip4 := []xlist.Resource{xlist.IPv4}
	mockup := &mockxl.List{ResourceList: ip4, Results: []bool{true}}

	var tests = []struct {
		reason string
		policy string
		want   string
	}{
		{"razon", "", "razon"},                                                                   //0
		{"razon", "[policy][/policy]", "razon"},                                                  //1
		{"razon", "[policy]kk=12[/policy]", "[policy]kk=12[/policy]razon"},                       //2
		{"ra[policy]kk=11[/policy]zon", "", "[policy]kk=11[/policy]razon"},                       //3
		{"razon[policy]kk=11[/policy]", "[policy]kk=12[/policy]", "[policy]kk=12[/policy]razon"}, //4
	}
	for idx, test := range tests {
		mockup.Reason = test.reason
		policy := reason.NewPolicy()
		policy.FromString(test.policy)
		checker := policywr.New(mockup, policy, policywr.Config{Merge: true})

		got, _ := checker.Check(context.Background(), "10.10.10.10", xlist.IPv4)
		if got.Reason != test.want {
			t.Errorf("idx[%v] policwr.Check(): want=%v got=%v", idx, test.want, got.Reason)
		}
	}
}

func TestWrapper_Threshold(t *testing.T) {
	ip4 := []xlist.Resource{xlist.IPv4}
	mockup := &mockxl.List{ResourceList: ip4, Results: []bool{true}}

	var tests = []struct {
		reason    string
		policy    string
		threshold int
		want      string
	}{
		{"razon", "[policy]kk=1[/policy]", 0, "razon"}, //0
		{"[score]1[/score]razon", "[policy]kk=1[/policy]", 0,
			"[policy]kk=1[/policy][score]1[/score]razon"}, //1
		{"razon", "[policy]kk=1[/policy]", -1, "[policy]kk=1[/policy]razon"}, //2
		{"razon[score]1[/score]", "[policy]kk=1[/policy]", 0,
			"[policy]kk=1[/policy]razon[score]1[/score]"}, //3
	}
	for idx, test := range tests {
		mockup.Reason = test.reason
		policy := reason.NewPolicy()
		policy.FromString(test.policy)
		checker := policywr.New(mockup, policy,
			policywr.Config{
				UseThreshold: true,
				Score:        test.threshold,
			})

		got, _ := checker.Check(context.Background(), "10.10.10.10", xlist.IPv4)
		if got.Reason != test.want {
			t.Errorf("idx[%v] policwr.Check(): want=%v got=%v", idx, test.want, got.Reason)
		}
	}
}
