// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package mockxl_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/xlist/pkg/xlistd/components/mockxl"
)

func TestMockup(t *testing.T) {
	var tests = []struct {
		//construction opts
		resources []xlist.Resource
		results   []bool
		fail      bool
		//query
		name     string
		resource xlist.Resource
		//expected
		want    bool
		wantErr bool
	}{
		{[]xlist.Resource{xlist.IPv4}, []bool{}, false,
			"10.10.10.", xlist.IPv4,
			false, true},
		{[]xlist.Resource{xlist.IPv4}, []bool{}, false,
			"10.10.10.10", xlist.IPv4,
			false, false},
		{[]xlist.Resource{xlist.Domain}, []bool{}, false,
			"10.10.10.10", xlist.IPv4,
			false, true},
		{[]xlist.Resource{xlist.IPv4}, []bool{true}, false,
			"10.10.10.10", xlist.IPv4,
			true, false},
		{[]xlist.Resource{xlist.IPv4}, []bool{true}, true,
			"10.10.10.10", xlist.IPv4,
			false, true},
	}
	for idx, test := range tests {
		mock := &mockxl.List{
			ResourceList: test.resources,
			Results:      test.results,
			Fail:         test.fail,
		}
		resp, err := mock.Check(context.Background(), test.name, test.resource)
		if test.wantErr && err == nil {
			t.Errorf("mock.Check idx[%v] expected error", idx)
		} else if !test.wantErr && err != nil {
			t.Errorf("mock.Check idx[%v] unexpected error: %v", idx, err)
		}
		if test.want != resp.Result {
			t.Errorf("mock.Check idx[%v] want=%v got=%v", idx, test.want, resp.Result)
		}
	}
}

func TestMockupResults(t *testing.T) {
	results := []bool{true, true, false, false}
	mock := &mockxl.List{
		ResourceList: []xlist.Resource{xlist.IPv4},
		Results:      results,
	}
	for i := 0; i < 2; i++ {
		for idx, result := range results {
			resp, err := mock.Check(context.Background(), "10.10.10.10", xlist.IPv4)
			if err != nil {
				t.Errorf("[%v] mock.Check idx[%v] unexpected error: %v", i, idx, err)
				continue
			}
			if result != resp.Result {
				t.Errorf("[%v] mock.Check idx[%v] want=%v got=%v", i, idx, result, resp.Result)
			}
		}
	}
}

func TestMockupLazy(t *testing.T) {
	mock := &mockxl.List{
		ResourceList: []xlist.Resource{xlist.IPv4},
		Results:      []bool{true},
		Lazy:         100 * time.Millisecond,
	}
	//test1
	ctx1, cancel1 := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel1()
	_, err := mock.Check(ctx1, "10.10.10.10", xlist.IPv4)
	if err != nil {
		t.Errorf("mock.Check idx[test1] unexpected error: %v", err)
	}
	//test2
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel2()
	_, err = mock.Check(ctx2, "10.10.10.10", xlist.IPv4)
	if err == nil {
		t.Error("mock.Check idx[test2] expected error")
	} else if err != xlist.ErrCanceledRequest {
		t.Errorf("mock.Check idx[test1] unexpected error: %v", err)
	}
}

func ExampleList() {
	// creating a list that fails all the time
	testlist := &mockxl.List{Fail: true}
	err := testlist.Ping()
	if err != nil {
		fmt.Println("ping 1:", err)
	}

	// creating a list that checks ipv4 resources and returns
	// a sequence of one postive result and two negative results
	testlist = &mockxl.List{
		ResourceList: []xlist.Resource{xlist.IPv4},
		Reason:       "hey, it's on the list",
		Results:      []bool{true, false, false},
	}

	err = testlist.Ping()
	if err != nil {
		log.Fatalln("this should not happen")
	}
	fmt.Println("ping 2:", err)

	resources := testlist.Resources()
	fmt.Println("resources:", resources)

	resp, err := testlist.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != nil || !resp.Result {
		log.Fatalln("this should not happen")
	}
	fmt.Println("check 1:", resp.Result, resp.Reason)

	resp, err = testlist.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != nil || resp.Result {
		log.Fatalln("this should not happen")
	}
	fmt.Println("check 2:", resp.Result)

	// now, we setup lazy responses...
	testlist.Lazy = 100 * time.Millisecond

	resp, err = testlist.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != nil || resp.Result {
		log.Fatalln("this should not happen")
	}
	fmt.Println("check 3:", resp.Result)

	// if we setup a timeout...
	timeoutCtx, cancelctx := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancelctx()

	resp, err = testlist.Check(timeoutCtx, "10.10.10.10", xlist.IPv4)
	if err == nil {
		log.Fatalln("this sould not nappen")
	}
	fmt.Println("check 4:", err)

	// Output:
	// ping 1: not available
	// ping 2: <nil>
	// resources: [ip4]
	// check 1: true hey, it's on the list
	// check 2: false
	// check 3: false
	// check 4: canceled request
}
