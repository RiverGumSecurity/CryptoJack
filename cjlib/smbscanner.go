package cjlib

import (
    "fmt"
    "github.com/stacktitan/smb/smb"
    "github.com/seancfoley/ipaddress-go/ipaddr"
    "log"
    "sync"
)

func PrimaryIPAddr() string {
    addrlist, _ := AddressList()
    for _, i := range addrlist {
        if ValidIPv4(i) {
            return i
        }
    }
    return ""
}

func SMBScanSubnet(username string, password string, domain string) {
    addr := PrimaryIPAddr()
    network := ipaddr.NewIPAddressString(addr).GetAddress().ToPrefixBlock()
    fmt.Printf("[*] SMB Subnet Scan Started for [%s] using [%s\\%s:%s]\n", network, domain, username, password)
    subnet := ipaddr.NewIPAddressString(addr).GetAddress().ToPrefixBlock().WithoutPrefixLen()
    iterator := subnet.Iterator()
    _ = iterator.Next() //throw away network addr
    i := 2
    var wg sync.WaitGroup
    for next := iterator.Next(); next != nil; next = iterator.Next() {
        if int64(i) >= subnet.GetCount().Int64() {
            break
        }
        wg.Add(1)
        go SMBScanTarget(next.String(), username, password, domain, &wg)
        i += 1
    }
    wg.Wait()
}

func SMBScanTarget(target string, username string, password string, domain string, wg *sync.WaitGroup) {
    defer wg.Done()
    options := smb.Options{
        Host:        target,
        Port:        445,
        User:        username,
        Domain:      domain,
        Workstation: "",
        Password:    password,
    }
    session, err := smb.NewSession(options, false)
    if err != nil {
        //log.Print("[!] ", err)
        return
    }
    defer session.Close()

    if session.IsAuthenticated {
        log.Print("[+] SMB Login successful to ", target)
    } else {
        log.Print("[-] SMB Login failed to ", target)
    }
    if session.IsSigningRequired {
        log.Print("[+] SMB Signing is required on ", target)
    } else {
        log.Print("[-] SMB Signing is NOT required on ", target)
    }
}
