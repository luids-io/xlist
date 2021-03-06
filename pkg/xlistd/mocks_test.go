// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package xlistd_test

import (
	"context"
	"fmt"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/xlist/pkg/xlistd"
)

type mockList struct {
	id        string
	fail      bool
	resources []xlist.Resource
	response  xlist.Response
	//lazy      time.Duration
}

func (l mockList) Check(ctx context.Context, name string, res xlist.Resource) (xlist.Response, error) {
	name, ctx, err := xlist.DoValidation(ctx, name, res, false)
	if err != nil {
		return xlist.Response{}, err
	}
	if !res.InArray(l.resources) {
		return xlist.Response{}, xlist.ErrNotSupported
	}
	if l.fail {
		return xlist.Response{}, xlist.ErrUnavailable
	}
	// if l.lazy > 0 {
	// 	time.Sleep(l.lazy)
	// }
	return l.response, nil
}

func (l mockList) ID() string {
	return l.id
}

func (l mockList) Class() string {
	return "mocktest"
}

func (l mockList) Ping() error {
	if l.fail {
		return xlist.ErrUnavailable
	}
	return nil
}

func (l mockList) Resources(ctx context.Context) ([]xlist.Resource, error) {
	ret := make([]xlist.Resource, len(l.resources), len(l.resources))
	copy(ret, l.resources)
	return ret, nil
}

func (l mockList) ReadOnly() bool {
	return false
}

type mockContainer struct {
	id          string
	stopOnError bool
	resources   []xlist.Resource
	lists       []xlistd.List
}

func (c mockContainer) ID() string {
	return c.id
}

func (c mockContainer) Class() string {
	return "mockcontainer"
}

func (c mockContainer) Check(ctx context.Context, name string, res xlist.Resource) (xlist.Response, error) {
	name, ctx, err := xlist.DoValidation(ctx, name, res, false)
	if err != nil {
		return xlist.Response{}, err
	}
	if !res.InArray(c.resources) {
		return xlist.Response{}, xlist.ErrNotSupported
	}
	for _, checker := range c.lists {
		resp, err := checker.Check(ctx, name, res)
		if err != nil && c.stopOnError {
			return xlist.Response{}, err
		}
		if resp.Result {
			return resp, err
		}
	}
	return xlist.Response{}, nil
}

func (c mockContainer) Ping() error {
	for _, checker := range c.lists {
		err := checker.Ping()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c mockContainer) Resources(ctx context.Context) ([]xlist.Resource, error) {
	var ret []xlist.Resource
	copy(ret, c.resources)
	return ret, nil
}

func (l mockContainer) ReadOnly() bool {
	return true
}

type mockWrapper struct {
	preffix string
	list    xlistd.List
}

func (w mockWrapper) ID() string {
	return w.list.ID()
}

func (w mockWrapper) Class() string {
	return w.list.Class()
}

func (w mockWrapper) Check(ctx context.Context, name string, res xlist.Resource) (xlist.Response, error) {
	resp, err := w.list.Check(ctx, name, res)
	resp.Reason = fmt.Sprintf("%s: %s", w.preffix, resp.Reason)
	return resp, err
}

func (w mockWrapper) Ping() error {
	return w.list.Ping()
}

func (w mockWrapper) Resources(ctx context.Context) ([]xlist.Resource, error) {
	return w.list.Resources(ctx)
}

func testBuilderList() xlistd.BuildListFn {
	return func(builder *xlistd.Builder, parents []string, list xlistd.ListDef) (xlistd.List, error) {
		response := xlist.Response{}
		if list.Source != "" {
			response.Result = true
			response.Reason = list.Source
		}
		bl := &mockList{
			response:  response,
			resources: xlist.ClearResourceDups(list.Resources, true),
		}
		if list.Opts != nil {
			fail, ok := list.Opts["fail"]
			if ok {
				fail, ok := fail.(bool)
				if !ok {
					return nil, fmt.Errorf("not valid value for 'fail' option in  %s list", list.ID)
				}
				bl.fail = fail
			}
			value, ok := list.Opts["ttl"]
			if ok {
				ttl, ok := value.(int)
				if !ok {
					//unmarshalling from json of some numbers as float64
					fttl, ok := value.(float64)
					if !ok {
						return nil, fmt.Errorf("not valid value for 'ttl' option in %s list", list.ID)
					}
					ttl = int(fttl)
				}
				bl.response.TTL = ttl
			}
		}
		return bl, nil
	}
}

func testBuilderCompo() xlistd.BuildListFn {
	return func(builder *xlistd.Builder, parents []string, list xlistd.ListDef) (xlistd.List, error) {
		if list.Contains == nil || len(list.Contains) == 0 {
			return nil, fmt.Errorf("no containers defined for %s list", list.ID)
		}

		bl := &mockContainer{resources: xlist.ClearResourceDups(list.Resources, true)}
		if list.Opts != nil {
			stopOpt, ok := list.Opts["stoponerror"]
			if ok {
				stopOpt, ok := stopOpt.(bool)
				if !ok {
					return nil, fmt.Errorf("not valid value for 'stoponerror' option in %s list", list.ID)
				}
				bl.stopOnError = stopOpt
			}
		}
		for _, sublist := range list.Contains {
			for _, r := range list.Resources {
				if !r.InArray(sublist.Resources) {
					return nil, fmt.Errorf("child %s doesn't checks resource %s", sublist.ID, r)
				}
			}
			slist, err := builder.BuildChild(append(parents, list.ID), sublist)
			if err != nil {
				return nil, err
			}
			bl.lists = append(bl.lists, slist)
		}
		return bl, nil
	}
}

func testBuilderWrap() xlistd.BuildWrapperFn {
	return func(builder *xlistd.Builder, def xlistd.WrapperDef, bl xlistd.List) (xlistd.List, error) {
		preffix := ""
		if def.Opts != nil {
			preffixs, ok := def.Opts["preffix"]
			if ok {
				preffixs, ok := preffixs.(string)
				if !ok {
					return nil, fmt.Errorf("can't get suffix value for response %s", bl.ID())
				}
				preffix = preffixs
			}
		}
		return &mockWrapper{preffix: preffix, list: bl}, nil
	}
}
