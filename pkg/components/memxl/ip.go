// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package memxl

import (
	"fmt"
	"net"

	"github.com/yl2chen/cidranger"

	"github.com/luids-io/core/xlist"
)

type ipList struct {
	//main hash list
	hashmap map[string]bool
	//lists of cidrs and subdomains
	ranger4 cidranger.Ranger
	ranger6 cidranger.Ranger
}

func newIPList() *ipList {
	l := &ipList{
		hashmap: make(map[string]bool),
		ranger4: cidranger.NewPCTrieRanger(),
		ranger6: cidranger.NewPCTrieRanger(),
	}
	return l
}

func (l *ipList) checkIP4(name string) bool {
	if l.inHashIP4(name) {
		return true
	}
	if l.inCIDR4s(name) {
		return true
	}
	return false
}

func (l *ipList) checkIP6(name string) bool {
	if l.inHashIP6(name) {
		return true
	}
	if l.inCIDR6s(name) {
		return true
	}
	return false
}

// funcions for add, warning, functions without lock!
func (l *ipList) addIP4(s string) error {
	k, ok := xlist.Canonicalize(s, xlist.IPv4)
	if !ok {
		return xlist.ErrBadRequest
	}
	key := fmt.Sprintf("ip4,%s", k)
	l.hashmap[key] = false
	return nil
}

func (l *ipList) removeIP4(s string) error {
	k, ok := xlist.Canonicalize(s, xlist.IPv4)
	if !ok {
		return xlist.ErrBadRequest
	}
	key := fmt.Sprintf("ip4,%s", k)
	delete(l.hashmap, key)
	return nil
}

func (l *ipList) addIP6(s string) error {
	k, ok := xlist.Canonicalize(s, xlist.IPv6)
	if !ok {
		return xlist.ErrBadRequest
	}
	key := fmt.Sprintf("ip6,%s", k)
	l.hashmap[key] = false
	return nil
}

func (l *ipList) removeIP6(s string) error {
	k, ok := xlist.Canonicalize(s, xlist.IPv6)
	if !ok {
		return xlist.ErrBadRequest
	}
	key := fmt.Sprintf("ip6,%s", k)
	delete(l.hashmap, key)
	return nil
}

func (l *ipList) addCIDR4(s string) error {
	ip, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return xlist.ErrBadRequest
	}
	if ip.To4() == nil {
		return xlist.ErrBadRequest
	}
	l.ranger4.Insert(cidranger.NewBasicRangerEntry(*ipnet))
	return nil
}

func (l *ipList) removeCIDR4(s string) error {
	ip, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return xlist.ErrBadRequest
	}
	if ip.To4() == nil {
		return xlist.ErrBadRequest
	}
	l.ranger4.Remove(*ipnet)
	return nil
}

func (l *ipList) addCIDR6(s string) error {
	ip, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return xlist.ErrBadRequest
	}
	if ip.To4() != nil {
		return xlist.ErrBadRequest
	}
	l.ranger6.Insert(cidranger.NewBasicRangerEntry(*ipnet))
	return nil
}

func (l *ipList) removeCIDR6(s string) error {
	ip, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return xlist.ErrBadRequest
	}
	if ip.To4() != nil {
		return xlist.ErrBadRequest
	}
	l.ranger6.Remove(*ipnet)
	return nil
}

// functions for check
func (l *ipList) inCIDR4s(ip string) bool {
	checkip := net.ParseIP(ip)
	ok, _ := l.ranger4.Contains(checkip)
	return ok
}

func (l *ipList) inCIDR6s(ip string) bool {
	checkip := net.ParseIP(ip)
	ok, _ := l.ranger6.Contains(checkip)
	return ok
}

func (l *ipList) inHashIP4(s string) bool {
	key := fmt.Sprintf("ip4,%s", s)
	_, ok := l.hashmap[key]
	return ok
}

func (l *ipList) inHashIP6(s string) bool {
	key := fmt.Sprintf("ip6,%s", s)
	_, ok := l.hashmap[key]
	return ok
}

// Clear internal data
func (l *ipList) clear() {
	l.hashmap = make(map[string]bool)
	l.ranger4 = cidranger.NewPCTrieRanger()
	l.ranger6 = cidranger.NewPCTrieRanger()
	//garbage collector has some work... ;)
}
