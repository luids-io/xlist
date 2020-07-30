// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package dnsxl

import (
	"net"
	"regexp"
	"strconv"
	"strings"
)

// note: we precompute for performance reasons
var validDomainRegexp, _ = regexp.Compile(`^(([a-zA-Z0-9]|[a-zA-Z0-9_][a-zA-Z0-9\-_]*[a-zA-Z0-9_])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)

func isDomain(s string) bool {
	return validDomainRegexp.MatchString(s)
}

func isIP(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil
}

func isValidPort(s string) bool {
	n, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	return (n > 0)
}

func reverse(name string) (rev string) {
	s := strings.Split(name, ".")
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	rev = strings.Join(s, ".")
	return
}

func ip6ToRecord(ip string) (record string) {
	split := strings.Split(ip, ":")
	//prepend with zerois if format ::
	prepend := make([]string, 0, 8)
	for i := len(split); i < 8; i++ {
		prepend = append(prepend, "0000")
	}
	split = append(prepend, split...)
	for idx := range split {
		if len(split[idx]) < 4 {
			split[idx] = strings.Repeat("0", 4-len(split[idx])) + split[idx]
		}
	}
	//mix all groups, split by char and uses dots as separator
	split = strings.Split(strings.Join(split, ""), "")
	record = strings.Join(split, ".")
	return
}
