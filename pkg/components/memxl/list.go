// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package memxl provides a xlist.Checker implementation that uses main
// memory for storage.
//
// This package is a work in progress and makes no API stability promises.
package memxl

import (
	"context"
	"sync"

	"github.com/luids-io/core/xlist"
)

// List stores all items in memory
type List struct {
	opts options
	//lock
	mu sync.RWMutex
	//resource lists
	iplist  *ipList
	domlist *domainList
	//resource types
	provides []bool
}

// Option is used for component configuration
type Option func(*options)

type options struct {
	forceValidation bool
	reason          string
}

var defaultOptions = options{}

// ForceValidation forces components to ignore context and validate requests
func ForceValidation(b bool) Option {
	return func(o *options) {
		o.forceValidation = b
	}
}

// Reason sets a fixed reason for component
func Reason(s string) Option {
	return func(o *options) {
		o.reason = s
	}
}

// New returns a new List
func New(resources []xlist.Resource, opt ...Option) *List {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	l := &List{
		opts:     opts,
		provides: make([]bool, len(xlist.Resources), len(xlist.Resources)),
	}
	//set resource types that provides
	for _, r := range xlist.ClearResourceDups(resources) {
		l.provides[int(r)] = true
	}
	l.init()
	return l
}

func (l *List) init() {
	if l.checks(xlist.IPv4) || l.checks(xlist.IPv6) {
		l.iplist = newIPList()
	}
	if l.checks(xlist.Domain) {
		l.domlist = newDomainList()
	}
}

// Check implements xlist.Checker interface
func (l *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if !l.checks(resource) {
		return xlist.Response{}, xlist.ErrResourceNotSupported
	}
	name, _, err := xlist.DoValidation(ctx, name, resource, l.opts.forceValidation)
	if err != nil {
		return xlist.Response{}, err
	}
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result bool
	switch resource {
	case xlist.IPv4:
		result = l.iplist.checkIP4(name)
	case xlist.IPv6:
		result = l.iplist.checkIP6(name)
	case xlist.Domain:
		result = l.domlist.checkDomain(name)
	}
	reason := ""
	if result {
		reason = l.opts.reason
	}
	return xlist.Response{Result: result, Reason: reason}, nil
}

// Resources implements xlist.Checker interface
func (l *List) Resources() []xlist.Resource {
	resources := make([]xlist.Resource, 0, len(xlist.Resources))
	for _, r := range xlist.Resources {
		if l.provides[int(r)] {
			resources = append(resources, r)
		}
	}
	return resources
}

// Ping implements xlist.Checker interface
func (l *List) Ping() error {
	return nil
}

// Add is a generic function for add a resource
func (l *List) Add(r xlist.Resource, f xlist.Format, s string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.add(r, f, s)
}

// generic funcion add resource, warning! no lock
func (l *List) add(r xlist.Resource, f xlist.Format, s string) error {
	if !l.checks(r) {
		return xlist.ErrResourceNotSupported
	}
	var err error
	switch r {
	case xlist.IPv4:
		switch f {
		case xlist.Plain:
			err = l.iplist.addIP4(s)
		case xlist.CIDR:
			err = l.iplist.addCIDR4(s)
		default:
			err = xlist.ErrBadResourceFormat
		}
	case xlist.IPv6:
		switch f {
		case xlist.Plain:
			err = l.iplist.addIP6(s)
		case xlist.CIDR:
			err = l.iplist.addCIDR6(s)
		default:
			err = xlist.ErrBadResourceFormat
		}
	case xlist.Domain:
		switch f {
		case xlist.Plain:
			err = l.domlist.addDomain(s)
		case xlist.Sub:
			err = l.domlist.addSubdomain(s)
		default:
			err = xlist.ErrBadResourceFormat
		}
	}
	return err
}

// AddIP4 add ip4
func (l *List) AddIP4(ip string) error {
	return l.AddIP4s([]string{ip})
}

// AddIP4s add a list of ip4
func (l *List) AddIP4s(ips []string) error {
	if !l.checks(xlist.IPv4) {
		return xlist.ErrResourceNotSupported
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, ip := range ips {
		err := l.iplist.addIP4(ip)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddIP6 add ip6
func (l *List) AddIP6(ip string) error {
	return l.AddIP6s([]string{ip})
}

// AddIP6s add a list of ip6
func (l *List) AddIP6s(ips []string) error {
	if !l.checks(xlist.IPv6) {
		return xlist.ErrResourceNotSupported
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, ip := range ips {
		err := l.iplist.addIP6(ip)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddCIDR4 adds CIDR
func (l *List) AddCIDR4(cidr string) error {
	return l.AddCIDR4s([]string{cidr})
}

// AddCIDR4s add a list of cidrs
func (l *List) AddCIDR4s(cidrs []string) error {
	if !l.checks(xlist.IPv4) {
		return xlist.ErrResourceNotSupported
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, cidr := range cidrs {
		err := l.iplist.addCIDR4(cidr)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddCIDR6 adds CIDR
func (l *List) AddCIDR6(cidr string) error {
	return l.AddCIDR6s([]string{cidr})
}

// AddCIDR6s add a list of cidrs
func (l *List) AddCIDR6s(cidrs []string) error {
	if !l.checks(xlist.IPv6) {
		return xlist.ErrResourceNotSupported
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, cidr := range cidrs {
		err := l.iplist.addCIDR6(cidr)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddDomain adds domain
func (l *List) AddDomain(domain string) error {
	return l.AddDomains([]string{domain})
}

// AddDomains add a list of domains
func (l *List) AddDomains(domains []string) error {
	if !l.checks(xlist.Domain) {
		return xlist.ErrResourceNotSupported
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, domain := range domains {
		err := l.domlist.addDomain(domain)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddSubdomain adds a subdomain
func (l *List) AddSubdomain(subdomain string) error {
	return l.AddSubdomains([]string{subdomain})
}

// AddSubdomains add a list of subdomains
func (l *List) AddSubdomains(subdomains []string) error {
	if !l.checks(xlist.Domain) {
		return xlist.ErrResourceNotSupported
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, subdomain := range subdomains {
		err := l.domlist.addSubdomain(subdomain)
		if err != nil {
			return err
		}
	}
	return nil
}

// Clear internal data
func (l *List) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.iplist != nil {
		l.iplist.clear()
	}
	if l.domlist != nil {
		l.domlist.clear()
	}
}

func (l *List) checks(r xlist.Resource) bool {
	if r >= xlist.IPv4 && r <= xlist.Domain {
		return l.provides[int(r)]
	}
	return false
}
