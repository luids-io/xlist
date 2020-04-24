// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package timeoutwr_test

import (
	"context"
	"testing"
	"time"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/components/mockxl"
	"github.com/luids-io/xlist/pkg/wrappers/timeoutwr"
)

func TestWrapper_Check(t *testing.T) {
	ip4 := []xlist.Resource{xlist.IPv4}

	t10ms := 10 * time.Millisecond
	t50ms := 50 * time.Millisecond
	t100ms := 100 * time.Millisecond

	mockup := &mockxl.List{
		ResourceList: ip4,
		Results:      []bool{true},
		Lazy:         t50ms,
	}

	var tests = []struct {
		timeout time.Duration
		wantErr bool
	}{
		{t100ms, false}, //0
		{t10ms, true},   //1
	}
	for idx, test := range tests {
		cache := timeoutwr.New(mockup, test.timeout)
		_, err := cache.Check(context.Background(), "10.10.10.10", xlist.IPv4)
		switch {
		case test.wantErr && err == nil:
			t.Errorf("idx[%v] cache.Check(): expected error", idx)
		case !test.wantErr && err != nil:
			t.Errorf("idx[%v] cache.Check(): err=%v", idx, err)
		}
	}
}
