// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package scorewr_test

import (
	"context"
	"testing"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/core/xlist/reason"
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
		checker := scorewr.New(mockup, score)

		got, _ := checker.Check(context.Background(), "10.10.10.10", xlist.IPv4)
		if got.Reason != test.want {
			t.Errorf("idx[%v] policwr.Check(): want=%v got=%v", idx, test.want, got.Reason)
		}
	}
}
