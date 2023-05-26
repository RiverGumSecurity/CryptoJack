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

func DNSLookup(name string, rectype string, extra ...string) ([]string, error) {
    var result []string
    var err error
    if rectype == "A" || rectype == "ANY" || rectype == "CNAME" {
        result, err = dnsResolver.LookupHost(context.Background(), name)
    } else if rectype == "PTR" {
        result, err = dnsResolver.LookupAddr(context.Background(), name)
    } else if rectype == "TXT" {
        result, err = dnsResolver.LookupTXT(context.Background(), name)
    } else if rectype == "SRV" {
        if len(extra) < 2 {
            return result, fmt.Errorf("not enough extra params")
        }
        var addr []*net.SRV
        _, addr, err = dnsResolver.LookupSRV(
            context.Background(), name, extra[0], extra[1])
        for _, s := range addr {
            result = append(result, fmt.Sprintf("%v:%v:%d:%d",
                s.Target, s.Port, s.Priority, s.Weight))
        }
    }
    return result, err
}
