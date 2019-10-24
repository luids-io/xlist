// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package ratewr_test

import (
	"context"
	"testing"
	"time"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/components/mockxl"
	"github.com/luids-io/xlist/pkg/wrappers/ratewr"
)

func TestWrapper_CheckNaive(t *testing.T) {
	ip4 := []xlist.Resource{xlist.IPv4}

	mockup := &mockxl.List{
		ResourceList: ip4,
		Results:      []bool{true},
	}

	var err error
	bucket := &ratewr.NaiveBucket{
		DripInterval: time.Second,
		Capacity:     10,
		PerDrip:      10,
	}
	bucket.Start()
	defer bucket.Shutdown()

	rated := ratewr.New(bucket, mockup)
	for i := 0; i < 10; i++ {
		_, err = rated.Check(context.Background(), "10.10.10.10", xlist.IPv4)
		if err != nil {
			t.Errorf("init rate error on %v: %v", i, err)
		}
	}

	_, err = rated.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err == nil {
		t.Fatalf("testing overrate: %v", err)
	}

	time.Sleep(time.Second + time.Millisecond)
	_, err = rated.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != nil {
		t.Errorf("rate error: %v", err)
	}
}
