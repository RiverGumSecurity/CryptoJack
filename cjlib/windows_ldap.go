package cjlib

import (
    "github.com/go-ldap/ldap"
    "net"
    "fmt"
    "time"
    "context"
)

func Find_LDAP_Server(domain string, ns string) error {
    r := &net.Resolver{}
    if len(ns) > 0 {
        r = &net.Resolver{
            PreferGo: true,
            Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
                d := net.Dialer{
                    Timeout: time.Millisecond * time.Duration(5000),
                }
                return d.DialContext(ctx, "udp", ns + ":53")
            },
        }
    }
    cname, addr, err := r.LookupSRV(context.Background(), "ldap._tcp.gc", "msdcs", domain)
    if err != nil {
        return err
    }
    fmt.Printf("\nCNAME: %s \n\n", cname)
    for _, srv := range addr {
        fmt.Printf("%v:%v:%d:%d\n", srv.Target, srv.Port, srv.Priority, srv.Weight)
    }
    return nil
}


func LDAPConnectBind(server string, username string, password string) (*ldap.Conn, error) {
    //server form: "ldap://domaincontroller.example.com:389"
    ldconn, err := ldap.Dial("tcp", server)
    if err != nil {
        return nil, err
    }
    defer ldconn.Close()
    err = ldconn.Bind(username, password)
    if err != nil {
        return nil, err
    }
    return ldconn, nil
}

func LDAPSearch(ldconn *ldap.Conn, baseDN string,
        filter string, attr []string) (*ldap.SearchResult, error) {
    //baseDN := "dc=example,dc=com"
    //filter := "(objectClass=user)"
    //attr := []string{"cn", "mail"}

    // Perform the LDAP search
    searchRequest := ldap.NewSearchRequest(
        baseDN,
        ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
        0, 0, false, filter, attr, nil,
    )

    searchResult, err := ldconn.Search(searchRequest)
    if err != nil {
        return nil, err
    }
    return searchResult, nil
}

