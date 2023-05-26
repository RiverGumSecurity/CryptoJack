package cjlib

import (
    "github.com/go-ldap/ldap"
    "fmt"
    "strings"
)

func FindLDAPServer(domain string) (string, error) {
    result, err := DNSLookup("ldaps._tcp.gc", "SRV", "msdcs", domain)
    if err != nil {
        result, err := DNSLookup("ldap._tcp.gc", "SRV", "msdcs", domain)
        if err != nil {
            return "", err
        }
        server := strings.TrimSuffix(strings.Split(result[0], ":")[0], ".")
        ip, err := DNSLookup(server, "A")
        if err != nil {
            fmt.Println(err.Error())
            return fmt.Sprintf("ldap://%s:389", server), nil
        }
        return fmt.Sprintf("ldap://%s:389", ip), nil
    }
    server := strings.TrimSuffix(strings.Split(result[0], ":")[0], ".")
    ip, err := DNSLookup(server, "A")
    if err != nil {
        fmt.Println(err.Error())
        return fmt.Sprintf("ldaps://%s:636", server), nil
    }
    return fmt.Sprintf("ldaps://%s:636", ip), nil
}

func _attr2Slice(attrs string) []string {
    var res []string
    for _, a := range strings.Split(attrs, ",") {
        res = append(res, strings.TrimSpace(a))
    }
    return res
}

func _baseDN(username string) string {
    domparts := strings.Split(strings.Split(username, "@")[1], ".")
    baseDN := ""
    for _, p := range domparts {
        baseDN += fmt.Sprintf("dc=%s,", p)
    }
    return strings.TrimSuffix(baseDN, ",")
}

func DomainUsers(ldap_server string, username string, password string) ([]string, error) {
    filter := "(&(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2))"
    return LDAPSearchWrapper(
        ldap_server, username, password,
        filter, "samAccountName")
}

func DomainComputers(ldap_server string, username string, password string) ([]string, error) {
    filter := "(&(samAccountType=805306369)(objectCategory=computer))"
    return LDAPSearchWrapper(
        ldap_server, username, password,
        filter, "samAccountName")
}

func LDAPSearchWrapper(ldap_server string, username string, password string, filter string, singleAttr string) ([]string, error) {
    ldapConn, err := LDAPConnectBind(ldap_server, username, password)
    if err != nil {
        return nil, err
    }
    defer ldapConn.Close()
    baseDN := _baseDN(username)
    var attrs = []string { singleAttr }
    searchResult, err:= LDAPSearch(ldapConn, baseDN, filter, attrs)
    if err != nil {
        return nil, err
    }
    var res []string
    for _, entry := range searchResult.Entries {
        res = append(res, entry.Attributes[0].Values[0])
    }
    return res, nil
}

func LDAPConnectBind(server string, username string, password string) (*ldap.Conn, error) {
    ldconn, err := ldap.DialURL(server)
    if err != nil {
        return nil, err
    }
    err = ldconn.Bind(username, password)
    if err != nil {
        return nil, err
    }
    return ldconn, nil
}

func LDAPSearch(ldconn *ldap.Conn, baseDN string,
        filter string, attr []string) (*ldap.SearchResult, error) {
    searchRequest := ldap.NewSearchRequest(
        baseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
        0, 0, false, filter, attr, nil,
    )
    searchResult, err := ldconn.Search(searchRequest)
    if err != nil {
        return nil, err
    }
    return searchResult, nil
}

