// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package dnsxl provides a xlist.Checker implementation that uses a dns zone
// as a source for its checks.
//
// This package is a work in progress and makes no API stability promises.
package dnsxl

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/yalogi"
)

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Logger:    yalogi.LogNull,
		DoReverse: true,
		Retries:   1,
		Timeout:   1 * time.Second,
	}
}

// Config options
type Config struct {
	Logger          yalogi.Logger
	Timeout         time.Duration
	ForceValidation bool
	DoReverse       bool
	HalfPing        bool
	ResolveReason   bool
	AuthToken       string
	Reason          string
	PingDNS         string
	Resolver        Resolver
	Retries         int
	DNSCodes        map[string]string
	ErrCodes        map[string]string
}

// Copy configuration
func (src Config) Copy() Config {
	dst := src
	if len(src.DNSCodes) > 0 {
		dst.DNSCodes = make(map[string]string, len(src.DNSCodes))
		for k, v := range src.DNSCodes {
			dst.DNSCodes[k] = v
		}
	}
	if len(src.ErrCodes) > 0 {
		dst.ErrCodes = make(map[string]string, len(src.ErrCodes))
		for k, v := range src.ErrCodes {
			dst.ErrCodes[k] = v
		}
	}
	return dst
}

//List implements an xlistd.List that checks against DNS blacklists
type List struct {
	id        string
	logger    yalogi.Logger
	opts      options
	client    *dns.Client
	resolver  Resolver
	zone      string
	dnsCodes  map[string]string //map reason codes
	errCodes  map[string]string //map error codes
	resources []xlist.Resource
	provides  []bool
}

type options struct {
	forceValidation bool
	doReverse       bool
	halfPing        bool
	resolveReason   bool
	authToken       string
	reason          string
	pingDNS         string
	retries         int
}

// New creates a new DNSxL based RBL
func New(id, zone string, resources []xlist.Resource, cfg Config) (*List, error) {
	if zone == "" {
		return nil, errors.New("zone parameter is required")
	}
	client := &dns.Client{}
	if cfg.Timeout > 0 {
		client.Timeout = cfg.Timeout
	}
	l := &List{
		id:     id,
		logger: cfg.Logger,
		opts: options{
			forceValidation: cfg.ForceValidation,
			doReverse:       cfg.DoReverse,
			halfPing:        cfg.HalfPing,
			resolveReason:   cfg.ResolveReason,
			authToken:       cfg.AuthToken,
			reason:          cfg.Reason,
			pingDNS:         cfg.PingDNS,
			retries:         cfg.Retries,
		},
		resolver: defaultResolver,
		client:   client,
		zone:     zone,
	}
	if cfg.Resolver != nil {
		l.resolver = cfg.Resolver
	}
	//set resource types that provides
	l.provides = make([]bool, len(xlist.Resources), len(xlist.Resources))
	l.resources = make([]xlist.Resource, 0, 3)
	for _, r := range xlist.ClearResourceDups(resources) {
		if r < xlist.IPv4 || r > xlist.Domain {
			return nil, fmt.Errorf("resource '%v' not supported", r)
		}
		l.provides[int(r)] = true
		l.resources = append(l.resources, r)
	}
	// set dns codes and error codes
	if len(cfg.DNSCodes) > 0 {
		l.dnsCodes = make(map[string]string, len(cfg.DNSCodes))
		for sip, reason := range cfg.DNSCodes {
			if !isIP(sip) {
				return l, fmt.Errorf("invalid ip '%s' in dnscodes", sip)
			}
			l.dnsCodes[sip] = reason
		}
	}
	if len(cfg.ErrCodes) > 0 {
		l.errCodes = make(map[string]string, len(cfg.ErrCodes))
		for sip, reason := range cfg.ErrCodes {
			if !isIP(sip) {
				return l, fmt.Errorf("invalid ip '%s' in errcodes", sip)
			}
			l.errCodes[sip] = reason
		}
	}
	return l, nil
}

// ID implements xlistd.List interface
func (l *List) ID() string {
	return l.id
}

// Class implements xlistd.List interface
func (l *List) Class() string {
	return BuildClass
}

// Check implements xlist.Checker interface
func (l *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if !l.checks(resource) {
		return xlist.Response{}, xlist.ErrNotSupported
	}
	name, ctx, err := xlist.DoValidation(ctx, name, resource, l.opts.forceValidation)
	if err != nil {
		return xlist.Response{}, err
	}
	//get resolver from pool
	server := l.resolver.Resolver()

	dnsrecord := name
	if resource == xlist.IPv6 {
		dnsrecord = ip6ToRecord(name)
	}
	if l.opts.doReverse {
		dnsrecord = reverse(dnsrecord)
	}
	if l.opts.authToken != "" {
		dnsrecord = fmt.Sprintf("%s.%s", l.opts.authToken, dnsrecord)
	}
	//get response
	resp, err := l.getResponse(ctx, server, dnsrecord)
	if err != nil {
		l.logger.Warnf("%s: check '%s': %v", l.id, name, err)
		return xlist.Response{}, xlist.ErrInternal
	}
	//if check and resolv reason...
	if resp.Result && l.opts.resolveReason {
		reason, err := l.getReason(ctx, server, dnsrecord)
		if err != nil {
			l.logger.Warnf("%s: get reason '%s': %v", l.id, name, err)
			reason = l.opts.reason
		}
		resp.Reason = reason
	}
	return resp, nil
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
	if l.opts.pingDNS != "" {
		return l.resolver.Ping(l.opts.pingDNS)
	}
	server := l.resolver.Resolver()
	for _, r := range l.Resources() {
		err := l.pingRFC5782(r, server)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadOnly implements xlistd.List interface
func (l *List) ReadOnly() bool {
	return true
}

//getResponse get response from A records
func (l *List) getResponse(ctx context.Context, server, dnsrecord string) (xlist.Response, error) {
	dnsquery := fmt.Sprintf("%s.%s.", dnsrecord, l.zone)
	r, err := l.queryDNS(ctx, server, dnsquery, dns.TypeA)
	if err != nil {
		return xlist.Response{}, err
	}
	if r.Rcode != dns.RcodeSuccess && r.Rcode != dns.RcodeNameError {
		return xlist.Response{}, fmt.Errorf("unexpected dns code %v(%v) in A query", r.Rcode, dns.RcodeToString[r.Rcode])
	}
	if r.Rcode == dns.RcodeNameError || (len(r.Answer) == 0) {
		return xlist.Response{}, nil //not in dns list
	}
	// process answer for get ttl and maps to list
	var ttl uint32
	reasons := []string{}
	for _, a := range r.Answer {
		if rsp, ok := a.(*dns.A); ok {
			//check if valid response
			resIP := rsp.A.String()
			if !strings.HasPrefix(resIP, "127.") {
				return xlist.Response{}, fmt.Errorf("response %s not valid", resIP)
			}
			//get data and maps to list
			if rsp.Hdr.Ttl > ttl {
				ttl = rsp.Hdr.Ttl
			}
			codeerror, ok := l.errCodes[resIP]
			if ok {
				return xlist.Response{}, errors.New(codeerror)
			}
			codereason, ok := l.dnsCodes[resIP]
			if ok {
				reasons = append(reasons, codereason)
			}
		}
	}
	// if fixed reason
	if l.opts.reason != "" {
		return xlist.Response{Result: true, Reason: l.opts.reason, TTL: int(ttl)}, nil
	}
	// mix reasons
	retReason := strings.Join(reasons, ";")
	return xlist.Response{Result: true, Reason: retReason, TTL: int(ttl)}, nil
}

//getReason get reason from TXT records
func (l *List) getReason(ctx context.Context, server, dnsrecord string) (string, error) {
	dnsquery := fmt.Sprintf("%s.%s.", dnsrecord, l.zone)
	r, err := l.queryDNS(ctx, server, dnsquery, dns.TypeTXT)
	if err != nil {
		return "", err
	}
	if r.Rcode != dns.RcodeSuccess {
		return "", fmt.Errorf("unexpected dns code %v(%v) in TXT query", r.Rcode, dns.RcodeToString[r.Rcode])
	}
	if len(r.Answer) == 0 {
		return "", nil
	}
	resp := []string{}
	for _, a := range r.Answer {
		if rsp, ok := a.(*dns.TXT); ok {
			resp = append(resp, strings.Join(rsp.Txt, " "))
		}
	}
	return strings.Join(resp, ";"), nil
}

func (l *List) checks(r xlist.Resource) bool {
	if r >= xlist.IPv4 && r <= xlist.Domain {
		return l.provides[int(r)]
	}
	return false
}

func (l *List) queryDNS(ctx context.Context, server, dnsquery string, dnstype uint16) (*dns.Msg, error) {
	m := &dns.Msg{}
	m.SetQuestion(dnsquery, dnstype)
	var r *dns.Msg
	var err error
	success := false
	count := 0
	for count < l.opts.retries && !success {
		select {
		case <-ctx.Done():
			return r, xlist.ErrCanceledRequest
		default:
		}
		r, _, err = l.client.ExchangeContext(ctx, m, server)
		if err == nil {
			success = true
		}
		count++
	}
	if err != nil {
		return r, fmt.Errorf("network problems with %v: %v", server, err)
	}
	return r, nil
}

func (l *List) pingRFC5782(r xlist.Resource, server string) error {
	switch r {
	case xlist.IPv4:
		return l.pingIPv4(server)
	case xlist.IPv6:
		return l.pingIPv6(server)
	case xlist.Domain:
		return l.pingDomain(server)
	}
	return xlist.ErrNotSupported
}

//implementation of RFC5782 checks
func (l *List) pingIPv4(server string) error {
	//127.0.0.2 must exists
	dnsrecord := "127.0.0.2"
	if l.opts.doReverse {
		dnsrecord = reverse(dnsrecord)
	}
	resp, err := l.getResponse(context.Background(), server, dnsrecord)
	if err != nil {
		return fmt.Errorf("dnsxl ping ipv4 failed: %v", err)
	}
	if !resp.Result {
		return fmt.Errorf("%v seems not be available: check 127.0.0.2 failed", l.zone)
	}
	if l.opts.halfPing {
		return nil
	}
	//127.0.0.1 must NOT exists
	dnsrecord = "127.0.0.1"
	if l.opts.doReverse {
		dnsrecord = reverse(dnsrecord)
	}
	resp, err = l.getResponse(context.Background(), server, dnsrecord)
	if err != nil {
		return fmt.Errorf("dnsxl ping ipv4 failed: %v", err)
	}
	if resp.Result {
		return fmt.Errorf("%v seems not be available: check 127.0.0.1 failed", l.zone)
	}
	return nil
}

func (l *List) pingIPv6(server string) error {
	//::FFFF:7F00:2 must exists
	dnsrecord := ip6ToRecord("::FFFF:7F00:2")
	if l.opts.doReverse {
		dnsrecord = reverse(dnsrecord)
	}
	resp, err := l.getResponse(context.Background(), server, dnsrecord)
	if err != nil {
		return fmt.Errorf("dnsxl ping ipv6 failed: %v", err)
	}
	if !resp.Result {
		return fmt.Errorf("%v seems not be available: check ::FFFF:7F00:2 failed", l.zone)
	}
	if l.opts.halfPing {
		return nil
	}
	//::FFFF:7F00:1 must NOT exists
	dnsrecord = ip6ToRecord("::FFFF:7F00:1")
	if l.opts.doReverse {
		dnsrecord = reverse(dnsrecord)
	}
	resp, err = l.getResponse(context.Background(), server, dnsrecord)
	if err != nil {
		return fmt.Errorf("dnsxl ping ipv6 failed: %v", err)
	}
	if resp.Result {
		return fmt.Errorf("%v seems not be available: check ::FFFF:7F00:1 failed", l.zone)
	}
	return nil
}

func (l *List) pingDomain(server string) error {
	//TEST must exists
	dnsrecord := "TEST"
	resp, err := l.getResponse(context.Background(), server, dnsrecord)
	if err != nil {
		return fmt.Errorf("dnsxl ping domain failed: %v", err)
	}
	if !resp.Result {
		return fmt.Errorf("%v seems not be available: check TEST domain failed", l.zone)
	}
	if l.opts.halfPing {
		return nil
	}
	//INVALID must NOT exists
	dnsrecord = "INVALID"
	resp, err = l.getResponse(context.Background(), server, dnsrecord)
	if err != nil {
		return fmt.Errorf("dnsxl ping domain failed: %v", err)
	}
	if resp.Result {
		return fmt.Errorf("%v seems not be available: check INVALID domain failed", l.zone)
	}
	return nil
}
