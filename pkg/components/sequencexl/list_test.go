// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package sequencexl_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/components/mockxl"
	"github.com/luids-io/xlist/pkg/components/sequencexl"
)

func TestList_Check(t *testing.T) {
	ip4 := []xlist.Resource{xlist.IPv4}

	rblFalse := &mockxl.List{ResourceList: ip4}
	rblTrue := &mockxl.List{ResourceList: ip4, Results: []bool{true}}
	rblFail := &mockxl.List{ResourceList: ip4, Fail: true}
	rblLazy := &mockxl.List{ResourceList: ip4, Lazy: 10 * time.Millisecond}

	var tests = []struct {
		resources []xlist.Resource
		sequence  []xlist.Checker
		timeout   time.Duration
		stoOnErr  bool
		want      bool
		wantErr   bool
	}{
		{ip4, []xlist.Checker{}, 0, true, false, false},
		{ip4, []xlist.Checker{rblFalse}, 0, true, false, false},
		{ip4, []xlist.Checker{rblTrue}, 0, true, true, false},
		{ip4, []xlist.Checker{rblFalse, rblTrue}, 0, true, true, false},
		{ip4, []xlist.Checker{rblTrue, rblFalse}, 0, true, true, false},
		{ip4, []xlist.Checker{rblFalse, rblFalse, rblFalse}, 0, true, false, false},
		{ip4, []xlist.Checker{rblFalse, rblFalse, rblTrue}, 0, true, true, false},
		// errors
		{[]xlist.Resource{xlist.Domain}, []xlist.Checker{}, 0, true, false, true},
		{ip4, []xlist.Checker{rblFalse, rblFail, rblTrue}, 0, true, false, true},
		{ip4, []xlist.Checker{rblFalse, rblFail, rblTrue}, 0, false, true, false},
		{ip4, []xlist.Checker{rblLazy, rblFalse, rblTrue},
			19 * time.Millisecond, true, true, false},
		{ip4, []xlist.Checker{rblLazy, rblLazy, rblTrue},
			19 * time.Millisecond, true, false, true},
		{ip4, []xlist.Checker{rblLazy, rblLazy, rblTrue},
			19 * time.Millisecond, false, false, true},
	}
	for idx, test := range tests {
		wseq := sequencexl.New(test.resources, sequencexl.SkipErrors(!test.stoOnErr), sequencexl.FirstResponse(true))
		for _, rbl := range test.sequence {
			wseq.AddChecker(rbl)
		}
		//create context with timeout
		ctx := context.Background()
		if test.timeout > 0 {
			var cancelctx context.CancelFunc
			ctx, cancelctx = context.WithTimeout(ctx, test.timeout)
			defer cancelctx()
		}
		//do the check
		resp, err := wseq.Check(ctx, "10.10.10.10", xlist.IPv4)
		if test.wantErr && err == nil {
			t.Errorf("sequence.Check idx[%v] expected error", idx)
		} else if !test.wantErr && err != nil {
			t.Errorf("sequence.Check idx[%v] unexpected error: %v", idx, err)
		}
		if test.want != resp.Result {
			t.Errorf("sequence.Check idx[%v] want=%v got=%v", idx, test.want, resp.Result)
		}
	}
}

func ExampleList() {
	ip4 := []xlist.Resource{xlist.IPv4}

	rbl1 := &mockxl.List{
		ResourceList: ip4,
		Results:      []bool{true, false},
		Reason:       "rbl1"}
	rbl2 := &mockxl.List{
		ResourceList: ip4,
		Fail:         true}
	rbl3 := &mockxl.List{
		ResourceList: ip4,
		Results:      []bool{true, false},
		Reason:       "rbl3"}
	rbl4 := &mockxl.List{
		ResourceList: ip4,
		Results:      []bool{true, false},
		Reason:       "rbl4"}

	//constructs sequence rbl
	rbl := sequencexl.New(ip4, sequencexl.SkipErrors(true), sequencexl.FirstResponse(true))
	rbl.AddChecker(rbl1)
	rbl.AddChecker(rbl2) //rbl2 allways fails, but with skiperrors ignores it
	rbl.AddChecker(rbl3)
	rbl.AddChecker(rbl4)

	for i := 1; i < 5; i++ {
		resp, err := rbl.Check(context.Background(), "10.10.10.10", xlist.IPv4)
		if err != nil {
			log.Fatalln("this should not happen")
		}
		fmt.Printf("check %v: %v %v\n", i, resp.Result, resp.Reason)
	}

	// iter 1 ->
	//     check rbl1 == true -> returns true; now rbl1=false
	// iter 2 ->
	//      check rbl1 == false; now rbl1=true
	//      check rbl3 == true -> returns true; now rbl3=false
	// iter 3 ->
	//      check rbl1 == true -> returns true; now rbl1=false
	// iter 4 ->
	//      check rbl1 == false; now rbl1= true
	//      check rbl3 == false; now rbl3=true;
	//      check rbl4 == true -> returns true; now rbl4=false

	// Output:
	// check 1: true rbl1
	// check 2: true rbl3
	// check 3: true rbl1
	// check 4: true rbl4
}
