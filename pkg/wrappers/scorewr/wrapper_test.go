// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package scorewr_test

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/api/xlist/reason"
	"github.com/luids-io/xlist/pkg/components/mockxl"
	"github.com/luids-io/xlist/pkg/wrappers/scorewr"
)

func TestWrapper_Score(t *testing.T) {
	ip4 := []xlist.Resource{xlist.IPv4}
	mockup := &mockxl.List{ResourceList: ip4, Results: []bool{true}}

	var tests = []struct {
		reason string
		score  string
		want   string
	}{
		{"razon", "", "razon"},                                                    //0
		{"razon", "[score][/score]", "razon"},                                     //1
		{"razon", "[score]12[/score]", "[score]12[/score]razon"},                  //2
		{"ra[score]11[/score]zon", "", "razon"},                                   //3
		{"razon[score]11[/score]", "[score]12[/score]", "[score]12[/score]razon"}, //4
	}
	for idx, test := range tests {
		mockup.Reason = test.reason
		score, _, _ := reason.ExtractScore(test.score)
		checker := scorewr.New(mockup, score, scorewr.Config{})

		got, _ := checker.Check(context.Background(), "10.10.10.10", xlist.IPv4)
		if got.Reason != test.want {
			t.Errorf("idx[%v] policwr.Check(): want=%v got=%v", idx, test.want, got.Reason)
		}
	}
}

func TestWrapper_ScoreWithRepexp(t *testing.T) {
	ip4 := []xlist.Resource{xlist.IPv4}
	mockup := &mockxl.List{ResourceList: ip4, Results: []bool{true}}

	malwareExpr, _ := regexp.Compile("malware")
	phisingExpr, _ := regexp.Compile("phising")
	dangerousExpr, _ := regexp.Compile("dangerous")
	expr1 := []scorewr.ScoreExpr{
		{RegExp: malwareExpr, Score: 20},
		{RegExp: phisingExpr, Score: 20},
		{RegExp: dangerousExpr, Score: 10},
	}

	var tests = []struct {
		reason string
		score  int
		expr   []scorewr.ScoreExpr
		want   string
	}{
		{"noesta", 10, expr1, "[score]10[/score]"},                                   //0
		{"this is phising", 10, expr1, "[score]20[/score]"},                          //1
		{"this is a dangerous phising", 10, expr1, "[score]30[/score]"},              //2
		{"this is malware", 10, expr1, "[score]20[/score]"},                          //3
		{"this is a dangerous malware", 10, expr1, "[score]30[/score]"},              //4
		{"this is a dangerous malware with phising", 10, expr1, "[score]50[/score]"}, //5
		// must ignore previous score
		{"[score]20[/score]this is a dangerous malware with phising", 10, expr1, "[score]50[/score]"}, //5
	}
	for idx, test := range tests {
		mockup.Reason = test.reason
		checker := scorewr.New(mockup, test.score, scorewr.Config{Scores: test.expr})

		got, _ := checker.Check(context.Background(), "10.10.10.10", xlist.IPv4)
		if !strings.HasPrefix(got.Reason, test.want) {
			t.Errorf("idx[%v] policwr.Check(): wantprefix=%v got=%v", idx, test.want, got.Reason)
		}
	}
}
