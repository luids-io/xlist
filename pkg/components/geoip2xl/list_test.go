// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package geoip2xl_test

import (
	"context"
	"testing"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/components/geoip2xl"
)

var testdb1 = "../../../test/testdata/GeoIP2-Country-Test.mmdb"

func TestList_New(t *testing.T) {
	//test non existdb
	geoip := geoip2xl.New("nonexistent.mmdb", geoip2xl.Rules{})
	err := geoip.Start()
	if err == nil {
		t.Errorf("geoip.Start(): expected error")
	}
	//test db
	geoip = geoip2xl.New(testdb1, geoip2xl.Rules{
		Countries: []string{"ES"}, Reverse: true,
	})
	//test before start
	err = geoip.Ping()
	if err != xlist.ErrListNotAvailable {
		t.Errorf("geoip.Ping(): err=%v", err)
	}
	_, err = geoip.Check(context.Background(), "10.10.10.10", xlist.IPv4)
	if err != xlist.ErrListNotAvailable {
		t.Errorf("geoip.Check(): err=%v", err)
	}
	// test start
	err = geoip.Start()
	if err != nil {
		t.Fatalf("geoip.Start(): err=%v", err)
	}
	defer geoip.Shutdown()

	err = geoip.Ping()
	if err != nil {
		t.Errorf("geoip.Ping(): err=%v", err)
	}
	//test unsupported checks
	_, err = geoip.Check(context.Background(), "www.google.com", xlist.Domain)
	if err != xlist.ErrResourceNotSupported {
		t.Errorf("geoip.Check(): err=%v", err)
	}
}

func TestList_Check(t *testing.T) {
	var tests = []struct {
		in   geoip2xl.Rules
		want bool
	}{
		{geoip2xl.Rules{}, false},
		{geoip2xl.Rules{Countries: []string{"ES"}}, false},
		{geoip2xl.Rules{Countries: []string{"GB"}}, true},
		{geoip2xl.Rules{Countries: []string{"gb"}}, true},
		{geoip2xl.Rules{Countries: []string{"ES", "FR"}}, false},
		{geoip2xl.Rules{Countries: []string{"ES", "GB"}}, true},
		{geoip2xl.Rules{Countries: []string{"ES", "FR"}, Reverse: true}, true},
		{geoip2xl.Rules{Countries: []string{"ES", "GB"}, Reverse: true}, false},
	}
	for idx, test := range tests {
		geoip := geoip2xl.New(testdb1, test.in)
		err := geoip.Start()
		if err != nil {
			t.Fatalf("geoip.Start(): err=%v", err)
		}
		got, err := geoip.Check(context.Background(), "81.2.69.160", xlist.IPv4)
		if err != nil {
			t.Errorf("idx[%v] geoip.Check(): err=%v", idx, err)
		}
		if got.Result != test.want {
			t.Errorf("idx[%v] geoip.Check(): want=%v got=%v", idx, test.want, got)
		}
		geoip.Shutdown()
	}
}
