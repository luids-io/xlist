// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package memxl

import (
	"github.com/luids-io/api/xlist"
)

type hashList struct {
	//main hash list
	hashmap map[string]bool
}

func newHashList() *hashList {
	l := &hashList{
		hashmap: make(map[string]bool),
	}
	return l
}

func (l *hashList) check(name string) bool {
	_, ok := l.hashmap[name]
	return ok
}

// funcions for add, warning, functions without lock!
func (l *hashList) addMD5(s string) error {
	k, ok := xlist.Canonicalize(s, xlist.MD5)
	if !ok {
		return xlist.ErrBadRequest
	}
	l.hashmap[k] = true
	return nil
}

func (l *hashList) removeMD5(s string) error {
	k, ok := xlist.Canonicalize(s, xlist.MD5)
	if !ok {
		return xlist.ErrBadRequest
	}
	delete(l.hashmap, k)
	return nil
}

func (l *hashList) addSHA1(s string) error {
	k, ok := xlist.Canonicalize(s, xlist.SHA1)
	if !ok {
		return xlist.ErrBadRequest
	}
	l.hashmap[k] = true
	return nil
}

func (l *hashList) removeSHA1(s string) error {
	k, ok := xlist.Canonicalize(s, xlist.SHA1)
	if !ok {
		return xlist.ErrBadRequest
	}
	delete(l.hashmap, k)
	return nil
}

func (l *hashList) addSHA256(s string) error {
	k, ok := xlist.Canonicalize(s, xlist.SHA256)
	if !ok {
		return xlist.ErrBadRequest
	}
	l.hashmap[k] = true
	return nil
}

func (l *hashList) removeSHA256(s string) error {
	k, ok := xlist.Canonicalize(s, xlist.SHA256)
	if !ok {
		return xlist.ErrBadRequest
	}
	delete(l.hashmap, k)
	return nil
}

// Clear internal data
func (l *hashList) clear() {
	l.hashmap = make(map[string]bool)
	//garbage collector has some work... ;)
}
