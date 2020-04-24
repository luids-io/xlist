// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package dnsxl

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/miekg/dns"
)

var defaultResolver Resolver

//DefaultResolver sets the default resolver for all new instances
func DefaultResolver(resolver Resolver) {
	defaultResolver = resolver
}

//Resolver is an interface for dns resolvers
type Resolver interface {
	Resolver() string
	Ping(domain string) error
}

//ResolverRRPool implements Resolver interface with a round robbin pool
type ResolverRRPool struct {
	resolvers []string
	last, len int
	mu        sync.Mutex
}

//NewResolverRRPool constructs a dns server pool with passed addresses.
func NewResolverRRPool(addresses []string) (*ResolverRRPool, error) {
	if len(addresses) == 0 {
		return nil, errors.New("cannot create an empty resolver pool")
	}
	r := make([]string, 0, len(addresses))
	for _, address := range addresses {
		var server string
		l := strings.Split(address, ":")
		switch len(l) {
		case 1:
			if !isIP(l[0]) {
				return nil, fmt.Errorf("invalid resolver format in %s", address)
			}
			server = fmt.Sprintf("%s:53", l[0])

		case 2:
			if !isIP(l[0]) || !isValidPort(l[1]) {
				return nil, fmt.Errorf("invalid resolver format in %s", address)
			}
			server = address
		default:
			return nil, fmt.Errorf("invalid resolver format in %s", address)
		}
		r = append(r, server)
	}

	return &ResolverRRPool{resolvers: r, len: len(r)}, nil
}

//Resolver implements interface returning the next server in the pool
func (p *ResolverRRPool) Resolver() string {
	if p.len == 1 {
		return p.resolvers[0]
	}
	p.mu.Lock()
	p.last++
	if p.last >= p.len {
		p.last = 0
	}
	defer p.mu.Unlock()
	return p.resolvers[p.last]
}

//Ping implements interface Resolver checking that all resolvers in the pool
//resolves the domain name passed as parameter. Returns an error if something
//is not working.
func (p *ResolverRRPool) Ping(domain string) error {
	//prepares query and dns client
	c := &dns.Client{}
	m := &dns.Msg{}
	m.SetQuestion(fmt.Sprintf("%s.", domain), dns.TypeA)
	m.RecursionDesired = true

	//queries all the servers in the pool
	for _, server := range p.resolvers {
		r, _, err := c.Exchange(m, server)
		if err != nil {
			return fmt.Errorf("network problems with %v: %v", server, err)
		}
		if r.Rcode != dns.RcodeSuccess {
			return fmt.Errorf("can't resolv with server %s, response code %v(%v) in A query", server, r.Rcode, dns.RcodeToString[r.Rcode])
		}
		if len(r.Answer) == 0 {
			return fmt.Errorf("server %s response with an empty answer", server)
		}
	}
	return nil
}

// NewResolverPoolFromZone queries a dns zone for its name servers and returns a resolver pool
func NewResolverPoolFromZone(zone string) (*ResolverRRPool, error) {
	client := &dns.Client{}
	server := defaultResolver.Resolver()

	nameservers, err := getNameservers(client, server, zone)
	if err != nil {
		return nil, fmt.Errorf("can't create resolver pool por zone %s: %v", zone, err)
	}
	if len(nameservers) == 0 {
		return nil, fmt.Errorf("can't create resolver pool for zone %s: no nameservers available", zone)
	}
	addresses := make([]string, 0)
	for _, ns := range nameservers {
		ips, err := lookup(client, server, ns)
		if err != nil {
			return nil, fmt.Errorf("can't create resolver pool for zone %s: %v", zone, err)
		}
		if len(ips) > 0 { //only catch the first one
			addresses = append(addresses, ips[0].String())
		}
	}
	pool, err := NewResolverRRPool(addresses)
	if err != nil {
		return nil, fmt.Errorf("can't create resolver pool for zone %s: %v", zone, err)
	}
	return pool, nil
}

// NewResolverFromConf returns a dns pool from a resolv.conf file
func NewResolverFromConf(file string) (*ResolverRRPool, error) {
	config, err := dns.ClientConfigFromFile(file)
	if err != nil {
		return nil, fmt.Errorf("parsing resolvconf: %v", err)
	}
	pool, err := NewResolverRRPool(config.Servers)
	if err != nil {
		return nil, fmt.Errorf("can't create resolver pool from file %s: %v", file, err)
	}
	return pool, nil
}

func getNameservers(client *dns.Client, server string, zone string) ([]string, error) {
	m := &dns.Msg{}
	m.SetQuestion(fmt.Sprintf("%s.", zone), dns.TypeNS)
	m.RecursionDesired = true
	r, _, err := client.Exchange(m, server)
	if err != nil {
		return []string{}, fmt.Errorf("can't get nameservers from zone %s: %v", zone, err)
	}
	if r.Rcode != dns.RcodeSuccess {
		return []string{}, fmt.Errorf("can't get nameservers from zone %s, response code %v(%v) in NS query", zone, r.Rcode, dns.RcodeToString[r.Rcode])
	}
	nameservers := make([]string, 0, len(r.Answer))
	for _, a := range r.Answer {
		if rsp, ok := a.(*dns.NS); ok {
			nameservers = append(nameservers, rsp.Ns)
		}
	}
	return nameservers, nil
}

func lookup(client *dns.Client, server string, name string) ([]net.IP, error) {
	m := &dns.Msg{}
	m.SetQuestion(dns.Fqdn(name), dns.TypeA)
	m.RecursionDesired = true
	r, _, err := client.Exchange(m, server)
	if err != nil {
		return []net.IP{}, fmt.Errorf("can't lookup name %s: %v", name, err)
	}
	if r.Rcode != dns.RcodeSuccess {
		return []net.IP{}, fmt.Errorf("can't lookup name %s, response code %v(%v) in A query", name, r.Rcode, dns.RcodeToString[r.Rcode])
	}
	ips := make([]net.IP, 0, len(r.Answer))
	for _, a := range r.Answer {
		if rsp, ok := a.(*dns.A); ok {
			ips = append(ips, rsp.A)
		}
	}
	return ips, nil
}

func init() {
	defaultResolver, _ = NewResolverRRPool([]string{"8.8.8.8:53"})
}
