// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package sequencexl_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/xlist/pkg/xlistd"
	"github.com/luids-io/xlist/pkg/xlistd/components/mockxl"
	"github.com/luids-io/xlist/pkg/xlistd/components/sequencexl"
)

func TestList_Check(t *testing.T) {
	rblFalse := &mockxl.List{ResourceList: onlyIPv4}
	rblTrue := &mockxl.List{ResourceList: onlyIPv4, Results: []bool{true}}
	rblFail := &mockxl.List{ResourceList: onlyIPv4, Fail: true}
	rblLazy := &mockxl.List{ResourceList: onlyIPv4, Lazy: 10 * time.Millisecond}

	var tests = []struct {
		resources []xlist.Resource
		sequence  []xlistd.List
		timeout   time.Duration
		stoOnErr  bool
		want      bool
		wantErr   bool
	}{
		{onlyIPv4, []xlistd.List{}, 0, true, false, false},
		{onlyIPv4, []xlistd.List{rblFalse}, 0, true, false, false},
		{onlyIPv4, []xlistd.List{rblTrue}, 0, true, true, false},
		{onlyIPv4, []xlistd.List{rblFalse, rblTrue}, 0, true, true, false},
		{onlyIPv4, []xlistd.List{rblTrue, rblFalse}, 0, true, true, false},
		{onlyIPv4, []xlistd.List{rblFalse, rblFalse, rblFalse}, 0, true, false, false},
		{onlyIPv4, []xlistd.List{rblFalse, rblFalse, rblTrue}, 0, true, true, false},
		// errors
		{[]xlist.Resource{xlist.Domain}, []xlistd.List{}, 0, true, false, true},
		{onlyIPv4, []xlistd.List{rblFalse, rblFail, rblTrue}, 0, true, false, true},
		{onlyIPv4, []xlistd.List{rblFalse, rblFail, rblTrue}, 0, false, true, false},
		{onlyIPv4, []xlistd.List{rblLazy, rblFalse, rblTrue}, 19 * time.Millisecond, true, true, false},
		{onlyIPv4, []xlistd.List{rblLazy, rblLazy, rblTrue}, 19 * time.Millisecond, true, false, true},
		{onlyIPv4, []xlistd.List{rblLazy, rblLazy, rblTrue}, 19 * time.Millisecond, false, false, true},
	}
	for idx, test := range tests {
		wseq := sequencexl.New("test", test.sequence, test.resources,
			sequencexl.Config{
				FirstResponse: true,
				SkipErrors:    !test.stoOnErr,
			})
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
	resources := []xlist.Resource{xlist.IPv4}

	childs := []xlistd.List{
		&mockxl.List{
			Identifier:   "rbl1",
			ResourceList: resources,
			Results:      []bool{true, false},
			Reason:       "rbl1",
		},
		&mockxl.List{
			Identifier:   "rbl2",
			ResourceList: resources,
			Fail:         true,
		},
		&mockxl.List{
			Identifier:   "rbl3",
			ResourceList: resources,
			Results:      []bool{true, false},
			Reason:       "rbl3",
		},
		&mockxl.List{
			Identifier:   "rbl4",
			ResourceList: resources,
			Results:      []bool{true, false},
			Reason:       "rbl4",
		},
	}

	//constructs sequence rbl
	rbl := sequencexl.New("test", childs, resources,
		sequencexl.Config{
			SkipErrors:    true,
			FirstResponse: true,
		})

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
