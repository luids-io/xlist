// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package memxl

import (
	"github.com/luids-io/core/xlist"
)

type domainList struct {
	//main hash list
	hashmap map[string]bool
	//lists subdomains
	minDepth, maxDepth int
}

func newDomainList() *domainList {
	l := &domainList{
		hashmap: make(map[string]bool),
	}
	return l
}

func (l *domainList) checkDomain(name string) bool {
	//no subdomains inserted, search in hash
	if l.maxDepth == 0 {
		_, ok := l.hashmap[name]
		return ok
	}
	// search in subdomains
	idx, depth := getDomainIdx(name)
	if depth <= l.minDepth {
		_, ok := l.hashmap[name]
		return ok
	}
	// exists subdomains less than domain depth
	i, j := l.minDepth, l.maxDepth
	if depth <= l.maxDepth {
		j = depth
	}
	for i <= j {
		subdomain := getSubdomain(i, name, idx)
		v, ok := l.hashmap[subdomain]
		if ok {
			if v { // v is true if its a subdomain
				return true
			}
			if depth == i {
				return true
			}
		}
		i++
	}
	// not found in subdomains and domain is greater than subdomain stored
	if depth > l.maxDepth {
		_, ok := l.hashmap[name]
		return ok
	}
	return false
}

// warning, functions without lock!
func (l *domainList) addDomain(s string) error {
	k, ok := xlist.Canonicalize(s, xlist.Domain)
	if !ok {
		return xlist.ErrBadRequest
	}
	//check if inserted
	_, ok = l.hashmap[k]
	if ok {
		//don't change value
		return nil
	}
	l.hashmap[k] = false
	return nil
}

func (l *domainList) removeDomain(s string) error {
	k, ok := xlist.Canonicalize(s, xlist.Domain)
	if !ok {
		return xlist.ErrBadRequest
	}
	//check if inserted
	value, ok := l.hashmap[k]
	if ok {
		if value {
			//don't remove, its a subdomain
			return nil
		}
		delete(l.hashmap, k)
	}
	return nil
}

func (l *domainList) addSubdomain(s string) error {
	k, ok := xlist.Canonicalize(s, xlist.Domain)
	if !ok {
		return xlist.ErrBadRequest
	}
	depth := getDepth(k)
	if l.maxDepth == 0 {
		l.minDepth = depth
		l.maxDepth = depth
	} else {
		if depth < l.minDepth {
			l.minDepth = depth
		}
		if depth > l.maxDepth {
			l.maxDepth = depth
		}
	}
	l.hashmap[k] = true
	return nil
}

func (l *domainList) removeSubdomain(s string) error {
	k, ok := xlist.Canonicalize(s, xlist.Domain)
	if !ok {
		return xlist.ErrBadRequest
	}
	//check if inserted
	value, ok := l.hashmap[k]
	if ok {
		if !value {
			//don't remove, its a domain
			return nil
		}
		delete(l.hashmap, k)
	}
	return nil
}

func (l *domainList) clear() {
	l.hashmap = make(map[string]bool)
	l.minDepth, l.maxDepth = 0, 0
	//garbage collector has some work... ;)
}

func getDepth(domain string) int {
	count := 1
	for _, char := range domain {
		if char == '.' {
			count++
		}
	}
	return count
}

func getDomainIdx(domain string) ([]int, int) {
	count := 1
	indexes := make([]int, 0)
	for pos, char := range domain {
		if char == '.' {
			count++
			indexes = append(indexes, pos)
		}
	}
	for i, j := 0, len(indexes)-1; i < j; i, j = i+1, j-1 {
		indexes[i], indexes[j] = indexes[j], indexes[i]
	}
	return indexes, count
}

func getSubdomain(depth int, domain string, indexes []int) string {
	if depth == 0 {
		return ""
	}
	if depth > len(indexes) {
		return domain
	}
	return domain[indexes[depth-1]+1:]
}
