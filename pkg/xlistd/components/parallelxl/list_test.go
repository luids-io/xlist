// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package parallelxl_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/xlist/pkg/xlistd"
	"github.com/luids-io/xlist/pkg/xlistd/components/mockxl"
	"github.com/luids-io/xlist/pkg/xlistd/components/parallelxl"
)

func TestList_Check(t *testing.T) {
	ip4 := []xlist.Resource{xlist.IPv4}

	t5ms := 5 * time.Millisecond
	t10ms := 10 * time.Millisecond
	t15ms := 15 * time.Millisecond

	rblFalse := &mockxl.List{ResourceList: ip4}
	rblTrue := &mockxl.List{ResourceList: ip4, Results: []bool{true}}
	rblFail := &mockxl.List{ResourceList: ip4, Fail: true}
	rblLazyF := &mockxl.List{ResourceList: ip4, Lazy: t10ms}
	rblLazyT := &mockxl.List{ResourceList: ip4, Lazy: t10ms, Results: []bool{true}}

	var tests = []struct {
		resources []xlist.Resource
		parallel  []xlistd.List
		timeout   time.Duration
		stoOnErr  bool

		want    bool
		wantErr bool
	}{
		{ip4, []xlistd.List{}, 0, true, false, false},                                 //0
		{ip4, []xlistd.List{rblFalse}, 0, true, false, false},                         //1
		{ip4, []xlistd.List{rblTrue}, 0, true, true, false},                           //2
		{ip4, []xlistd.List{rblFalse, rblTrue}, 0, true, true, false},                 //3
		{ip4, []xlistd.List{rblTrue, rblFalse}, 0, true, true, false},                 //4
		{ip4, []xlistd.List{rblFalse, rblFalse, rblFalse}, 0, true, false, false},     //5
		{ip4, []xlistd.List{rblFalse, rblFalse, rblTrue}, 0, true, true, false},       //6
		{ip4, []xlistd.List{rblLazyF, rblFalse, rblLazyF}, t15ms, true, false, false}, //7
		{ip4, []xlistd.List{rblLazyF, rblLazyF, rblTrue}, t5ms, true, true, false},    //8
		// errors
		{[]xlist.Resource{xlist.Domain}, []xlistd.List{}, 0, true, false, true},     //9
		{ip4, []xlistd.List{rblLazyF, rblFail, rblLazyF}, 0, true, false, true},     //10
		{ip4, []xlistd.List{rblFalse, rblFail, rblTrue}, 0, false, true, false},     //11
		{ip4, []xlistd.List{rblLazyF, rblFail, rblLazyT}, 0, true, false, true},     //12
		{ip4, []xlistd.List{rblLazyT, rblLazyF, rblLazyT}, t5ms, true, false, true}, //13
	}
	for idx, test := range tests {
		wpar := parallelxl.New("test", test.parallel, test.resources,
			parallelxl.Config{
				SkipErrors:    !test.stoOnErr,
				FirstResponse: true,
			})
		//create context with timeout
		ctx := context.Background()
		if test.timeout > 0 {
			var cancelctx context.CancelFunc
			ctx, cancelctx = context.WithTimeout(ctx, test.timeout)
			defer cancelctx()
		}
		//do the check
		resp, err := wpar.Check(ctx, "10.10.10.10", xlist.IPv4)
		if test.wantErr && err == nil {
			t.Errorf("parallel.Check idx[%v] expected error", idx)
		} else if !test.wantErr && err != nil {
			t.Errorf("parallel.Check idx[%v] unexpected error: %v", idx, err)
		}
		if test.want != resp.Result {
			t.Errorf("parallel.Check idx[%v] want=%v got=%v", idx, test.want, resp.Result)
		}
	}
}

func ExampleList() {
	ip4 := []xlist.Resource{xlist.IPv4}
	t5ms := 5 * time.Millisecond

	cfg := parallelxl.Config{
		SkipErrors:    true,
		FirstResponse: true,
	}
	childs := []xlistd.List{
		&mockxl.List{Results: []bool{false}, ResourceList: ip4, Reason: "rbl1"},
		&mockxl.List{Results: []bool{true, false}, ResourceList: ip4, Reason: "rbl2"},
		&mockxl.List{Results: []bool{true}, Lazy: t5ms, ResourceList: ip4, Reason: "rbl3"},
		//rbl4 allways fails, but with skiperrors ignores this fail
		&mockxl.List{Fail: true, ResourceList: ip4},
	}

	//constructs parallel rbl
	rbl := parallelxl.New("test", childs, ip4, cfg)
	for i := 0; i < 4; i++ {
		resp, err := rbl.Check(context.Background(), "10.10.10.10", xlist.IPv4)
		if err != nil {
			log.Fatalln("this should not happen")
		}
		fmt.Printf("%v %v\n", resp.Result, resp.Reason)
	}

	// Output:
	// true rbl2
	// true rbl3
	// true rbl2
	// true rbl3
}
