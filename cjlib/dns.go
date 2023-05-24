package cjlib

import (
    "net"
    "fmt"
    "time"
    "context"
)

var dnsResolver = &net.Resolver{}
var dnsTimeout = 5000

func SetDNSResolver(ns string) {
    dnsResolver = &net.Resolver{
        PreferGo: true,
        Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
            d := net.Dialer{
                Timeout: time.Millisecond * time.Duration(dnsTimeout),
            }
            return d.DialContext(ctx, "udp", ns + ":53")
        },
    }
}

func DNSLookup(name string, rectype string) ([]string, error) {
    var result []string
    if rectype == "A" || rectype == "ANY" {
        result, err := net.LookupHost(name)
    } else if rectype == "SRV" {
        _, addr, err := dnsResolver.LookupSRV(
            context.Background(), "ldap._tcp.gc", "msdcs", name)
        for _, s := range addr {
            result = append(result, fmt.Sprintf("%v:%v:%d:%d",
                s.Target, s.Port, s.Priority, s.Weight))
        }
    }
    return result, err
}
