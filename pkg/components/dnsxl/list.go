// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package dnsxl provides a xlist.Checker implementation that uses a dns zone
// as a source for its checks.
//
// This package is a work in progress and makes no API stability promises.
package dnsxl

import (
	"context"
	"fmt"
	"strings"

	"github.com/miekg/dns"

	"github.com/luids-io/core/xlist"
)

// Client is the struct that implements dns client.
// It's an miekg/dns.Client alias.
type Client = dns.Client

//List implements an xlist.Checker that checks against DNS blacklists
type List struct {
	opts     options
	client   *Client
	resolver Resolver
	zone     string
	dnsCodes map[string]string //map reason codes
	errCodes map[string]string //map error codes
	provides []bool            //resources
}

type options struct {
	forceValidation bool
	doReverse       bool
	halfPing        bool
	resolveReason   bool
	authToken       string
	reason          string
	pingDNS         string
	resolver        Resolver
	retries         int
}

var defaultOptions = options{
	doReverse: true,
	retries:   1,
}

// Option is used for common component configuration
type Option func(*options)

// ForceValidation forces component to ignore context and validate requests
func ForceValidation(b bool) Option {
	return func(o *options) {
		o.forceValidation = b
	}
}

//Reverse option reverses ips and domains for dns searchs
func Reverse(b bool) Option {
	return func(o *options) {
		o.doReverse = b
	}
}

//ResolveReason option enables to query TXT records
func ResolveReason(b bool) Option {
	return func(o *options) {
		o.resolveReason = b
	}
}

//Reason option fixes a default reason
func Reason(s string) Option {
	return func(o *options) {
		o.reason = s
	}
}

//UseDNSPing option is for lists that doesn't support rfc and replaces lookup
func UseDNSPing(query string) Option {
	return func(o *options) {
		o.pingDNS = query
	}
}

//HalfPing option is for lists that doesn't support fully rfc
func HalfPing(b bool) Option {
	return func(o *options) {
		o.halfPing = b
	}
}

//AuthToken option prepends token to dns queries
func AuthToken(token string) Option {
	return func(o *options) {
		o.authToken = token
	}
}

//UseResolver option sets a custom resolver for the list
func UseResolver(r Resolver) Option {
	return func(o *options) {
		o.resolver = r
	}
}

//Retries option sets the number of retries in dns queries
func Retries(i int) Option {
	return func(o *options) {
		o.retries = i
	}
}

// New creates a new DNSxL based RBL
func New(client *Client, zone string, resources []xlist.Resource, opt ...Option) *List {
	if zone == "" {
		panic("zone parameter is required")
	}
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	l := &List{
		opts:     opts,
		client:   client,
		resolver: defaultResolver,
		zone:     zone,
		dnsCodes: make(map[string]string),
		errCodes: make(map[string]string),
		provides: make([]bool, len(xlist.Resources), len(xlist.Resources)),
	}
	//set resource types that provides
	for _, r := range resources {
		if r >= xlist.IPv4 && r <= xlist.Domain {
			l.provides[int(r)] = true
		}
	}
	if opts.resolver != nil {
		l.resolver = opts.resolver
	}
	return l
}

// AddDNSCode adds an ip with its reason
func (l *List) AddDNSCode(ip string, reason string) {
	if !isIP(ip) {
		return // do nothing
	}
	l.dnsCodes[ip] = reason
}

// AddErrCode adds an ip as error
func (l *List) AddErrCode(ip string, msgerr string) {
	if !isIP(ip) {
		return // do nothing
	}
	l.errCodes[ip] = msgerr
}

// Check implements xlist.Checker interface
func (l *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if !l.checks(resource) {
		return xlist.Response{}, xlist.ErrResourceNotSupported
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
		return xlist.Response{}, fmt.Errorf("checking %s: %v", name, err)
	}
	//if check and resolv reason...
	if resp.Result && l.opts.resolveReason {
		reason, err := l.getReason(ctx, server, dnsrecord)
		if err != nil {
			reason = fmt.Sprintf("error getting reason: %v", err)
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
				return xlist.Response{}, fmt.Errorf("%s", codeerror)
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
			return r, ctx.Err()
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
	return xlist.ErrResourceNotSupported
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
