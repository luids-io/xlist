// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package memxl provides a xlist.Checker implementation that uses main
// memory for storage.
//
// This package is a work in progress and makes no API stability promises.
package memxl

import (
	"context"
	"sync"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// Config options
type Config struct {
	ForceValidation bool
	Reason          string
}

type options struct {
	forceValidation bool
	reason          string
}

// List stores all items in memory
type List struct {
	id   string
	opts options
	//lock
	mu sync.RWMutex
	//resource lists
	iplist   *ipList
	domlist  *domainList
	hashlist *hashList
	//resource types
	provides  []bool
	resources []xlist.Resource
}

// New returns a new List
func New(id string, resources []xlist.Resource, cfg Config) *List {
	l := &List{
		id: id,
		opts: options{
			forceValidation: cfg.ForceValidation,
			reason:          cfg.Reason,
		},
		resources: xlist.ClearResourceDups(resources),
		provides:  make([]bool, len(xlist.Resources), len(xlist.Resources)),
	}
	//set resource types that providess
	for _, r := range l.resources {
		l.provides[int(r)] = true
	}
	l.init()
	return l
}

// ID implements xlistd.List interface
func (l *List) ID() string {
	return l.id
}

// Class implements xlistd.List interface
func (l *List) Class() string {
	return BuildClass
}

func (l *List) init() {
	if l.checks(xlist.IPv4) || l.checks(xlist.IPv6) {
		l.iplist = newIPList()
	}
	if l.checks(xlist.Domain) {
		l.domlist = newDomainList()
	}
	if l.checks(xlist.MD5) || l.checks(xlist.SHA1) || l.checks(xlist.SHA256) {
		l.hashlist = newHashList()
	}
}

// Check implements xlist.Checker interface
func (l *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if !l.checks(resource) {
		return xlist.Response{}, xlist.ErrNotSupported
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
	case xlist.MD5, xlist.SHA1, xlist.SHA256:
		result = l.hashlist.check(name)
	}
	reason := ""
	if result {
		reason = l.opts.reason
	}
	return xlist.Response{Result: result, Reason: reason}, nil
}

// Resources implements xlist.Checker interface
func (l *List) Resources() []xlist.Resource {
	resources := make([]xlist.Resource, len(l.resources), len(l.resources))
	copy(resources, l.resources)
	return resources
}

// Ping implements xlist.Checker interface
func (l *List) Ping() error {
	return nil
}

// generic funcion add resource, warning! no lock
func (l *List) add(r xlist.Resource, f xlistd.Format, s string) error {
	if !l.checks(r) {
		return xlist.ErrNotSupported
	}
	var err error
	switch r {
	case xlist.IPv4:
		switch f {
		case xlistd.Plain:
			err = l.iplist.addIP4(s)
		case xlistd.CIDR:
			err = l.iplist.addCIDR4(s)
		default:
			err = xlist.ErrBadRequest
		}
	case xlist.IPv6:
		switch f {
		case xlistd.Plain:
			err = l.iplist.addIP6(s)
		case xlistd.CIDR:
			err = l.iplist.addCIDR6(s)
		default:
			err = xlist.ErrBadRequest
		}
	case xlist.Domain:
		switch f {
		case xlistd.Plain:
			err = l.domlist.addDomain(s)
		case xlistd.Sub:
			err = l.domlist.addSubdomain(s)
		default:
			err = xlist.ErrBadRequest
		}
	case xlist.MD5:
		switch f {
		case xlistd.Plain:
			err = l.hashlist.addMD5(s)
		default:
			err = xlist.ErrBadRequest
		}
	case xlist.SHA1:
		switch f {
		case xlistd.Plain:
			err = l.hashlist.addSHA1(s)
		default:
			err = xlist.ErrBadRequest
		}
	case xlist.SHA256:
		switch f {
		case xlistd.Plain:
			err = l.hashlist.addSHA256(s)
		default:
			err = xlist.ErrBadRequest
		}
	}
	return err
}

// generic funcion add resource, warning! no lock
func (l *List) remove(r xlist.Resource, f xlistd.Format, s string) error {
	if !l.checks(r) {
		return xlist.ErrNotSupported
	}
	var err error
	switch r {
	case xlist.IPv4:
		switch f {
		case xlistd.Plain:
			err = l.iplist.removeIP4(s)
		case xlistd.CIDR:
			err = l.iplist.removeCIDR4(s)
		default:
			err = xlist.ErrBadRequest
		}
	case xlist.IPv6:
		switch f {
		case xlistd.Plain:
			err = l.iplist.removeIP6(s)
		case xlistd.CIDR:
			err = l.iplist.removeCIDR6(s)
		default:
			err = xlist.ErrBadRequest
		}
	case xlist.Domain:
		switch f {
		case xlistd.Plain:
			err = l.domlist.removeDomain(s)
		case xlistd.Sub:
			err = l.domlist.removeSubdomain(s)
		default:
			err = xlist.ErrBadRequest
		}
	case xlist.MD5:
		switch f {
		case xlistd.Plain:
			err = l.hashlist.removeMD5(s)
		default:
			err = xlist.ErrBadRequest
		}
	case xlist.SHA1:
		switch f {
		case xlistd.Plain:
			err = l.hashlist.removeSHA1(s)
		default:
			err = xlist.ErrBadRequest
		}
	case xlist.SHA256:
		switch f {
		case xlistd.Plain:
			err = l.hashlist.removeSHA256(s)
		default:
			err = xlist.ErrBadRequest
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
		return xlist.ErrNotSupported
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
		return xlist.ErrNotSupported
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
		return xlist.ErrNotSupported
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
		return xlist.ErrNotSupported
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
		return xlist.ErrNotSupported
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
		return xlist.ErrNotSupported
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

// AddMD5 add md5 hash
func (l *List) AddMD5(hash string) error {
	return l.AddMD5s([]string{hash})
}

// AddMD5s add a list of md5
func (l *List) AddMD5s(hashes []string) error {
	if !l.checks(xlist.MD5) {
		return xlist.ErrNotSupported
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, hash := range hashes {
		err := l.hashlist.addMD5(hash)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddSHA1 add sha1 hash
func (l *List) AddSHA1(hash string) error {
	return l.AddSHA1s([]string{hash})
}

// AddSHA1s add a list of sha1 hashes
func (l *List) AddSHA1s(hashes []string) error {
	if !l.checks(xlist.SHA1) {
		return xlist.ErrNotSupported
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, hash := range hashes {
		err := l.hashlist.addSHA1(hash)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddSHA256 add sha256 hash
func (l *List) AddSHA256(hash string) error {
	return l.AddSHA256s([]string{hash})
}

// AddSHA256s add a list of sha256 hashes
func (l *List) AddSHA256s(hashes []string) error {
	if !l.checks(xlist.SHA256) {
		return xlist.ErrNotSupported
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, hash := range hashes {
		err := l.hashlist.addSHA256(hash)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *List) checks(r xlist.Resource) bool {
	if r >= xlist.IPv4 && r <= xlist.SHA256 {
		return l.provides[int(r)]
	}
	return false
}

// Append item
func (l *List) Append(ctx context.Context, name string, r xlist.Resource, f xlistd.Format) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.add(r, f, name)
}

// Remove item
func (l *List) Remove(ctx context.Context, name string, r xlist.Resource, f xlistd.Format) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.remove(r, f, name)
}

// Clear list
func (l *List) Clear() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.iplist != nil {
		l.iplist.clear()
	}
	if l.domlist != nil {
		l.domlist.clear()
	}
	if l.hashlist != nil {
		l.hashlist.clear()
	}
	return nil
}

// ReadOnly implements xlistd.List interface
func (l *List) ReadOnly() bool {
	return false
}
