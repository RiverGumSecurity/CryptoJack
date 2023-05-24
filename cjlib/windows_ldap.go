package cjlib

import (
    "github.com/go-ldap/ldap"
)

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

