package cjlib

import (
    "fmt"
    "github.com/stacktitan/smb/smb"
    "github.com/seancfoley/ipaddress-go/ipaddr"
    "github.com/hirochachacha/go-smb2"
    "log"
    "net"
    "sync"
    "time"
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
        //SMBScanTarget(next.String(), username, password, domain, &wg)
        go EnumerateShares(next.String(), username, password, &wg)
        i += 1
    }
    wg.Wait()
}

func SMBScanTarget(target string, username string, password string, domain string, wg *sync.WaitGroup) {
    defer wg.Done()
    options := smb.Options{
        Host: target,
        Port: 445,
        User: username,
        Domain: domain,
        Workstation: "",
        Password: password,
    }
    session, err := smb.NewSession(options, false)
    if err != nil {
        log.Print("[!] ", err)
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

func EnumerateShares(host string, username string, password string,  wg *sync.WaitGroup) error {
    defer wg.Done()
    host_and_port := fmt.Sprintf("%s:%d", host, 445)
    conn, err := net.DialTimeout("tcp", host_and_port, time.Second * 5)
    if err != nil { return err }
    defer conn.Close()
    d := &smb2.Dialer{
        Initiator: &smb2.NTLMInitiator{
            User: username,
            Password: password,
        },
    }
    s, err := d.Dial(conn)
    if err != nil { return err }
    defer s.Logoff()
    names, err := s.ListSharenames()
    if err != nil { return err }
    for _, name := range names {
        fmt.Printf("\\\\%s\\%s\n", host, name)
    }
    return nil
}
