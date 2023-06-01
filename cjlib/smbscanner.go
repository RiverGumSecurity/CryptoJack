package cjlib

import (
    "fmt"
    //"github.com/stacktitan/smb/smb"
    "golang.org/x/sys/windows"
    "github.com/hirochachacha/go-smb2"
    //"github.com/hectane/go-acl"
    "log"
    "net"
    "sync"
    "time"
    "strings"
    //"syscall"
    "unsafe"
)


type AclSizeInformation struct {
    AceCount      uint32
    AclBytesInUse uint32
    AclBytesFree  uint32
}

type Acl struct {
    AclRevision uint8
    Sbz1        uint8
    AclSize     uint16
    AceCount    uint16
    Sbz2        uint16
}

type AccessAllowedAce struct {
    AceType    uint8
    AceFlags   uint8
    AceSize    uint16
    AccessMask uint32
    SidStart   uint32
}

func PrimaryIPAddr() string {
    addrlist, _ := AddressList()
    for _, i := range addrlist {
        if ValidIPv4(i) {
            return i
        }
    }
    return ""
}

func SMBScanDomainComputers(username string, password string, domain string) error {
    ldap_server, err := FindLDAPServer(domain)
    if err != nil { return err }
    usernameDomain := fmt.Sprintf("%s@%s", username, domain)
    domainComputers, err := DomainComputers(ldap_server, usernameDomain, password)
    if err != nil { return err }
    fmt.Println(`
[*] ======================
[+]  Domain Computer List
[*] ======================`)

    ch := make(chan string)
    var wg sync.WaitGroup
    wg.Add(len(domainComputers))
    for _, c := range domainComputers {
        addrs, _ := DNSLookup(fmt.Sprintf("%s.%s", strings.TrimSuffix(c, "$"), domain), "A")
        ip := addrs[0]
        fmt.Printf("[+] %-12s (%s)\n", c, ip)
        go EnumerateShares(ip, c, username, password, ch, &wg)
    }
    go func() {
        wg.Wait()
        close(ch)
    }()

    fmt.Println(`
[*] ===============================
[*]  Domain Wide Share Enumeration
[*] ===============================`)
    for m := range(ch) {
        fmt.Println(m)
        GetDACL(m)
    }
    return nil
}

/*func SMBScanSubnet(username string, password string, domain string) {
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
        go EnumerateShares(next.String(), username, password, &wg)
        i += 1
    }
    wg.Wait()
}

func SMBLogin(target string, username string, password string, domain string, wg *sync.WaitGroup) {
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
*/

func EnumerateShares(ip string, hostname string, username string, password string, ch chan<-string, wg *sync.WaitGroup) error {
    defer wg.Done()
    ip_port := fmt.Sprintf("%s:%d", ip, 445)
    conn, err := net.DialTimeout("tcp", ip_port, time.Second * 5)
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
        ch <- fmt.Sprintf("[+] \\\\%s\\%s", hostname, name)
    }
    return nil
}

func GetDACL(path string) {
    advapi32 := windows.NewLazySystemDLL("advapi32.dll")
    GetNamedSecurityInfo := advapi32.NewProc("GetNamedSecurityInfoW")
    GetSecurityDescriptorDacl := advapi32.NewProc("GetSecurityDescriptorDacl")
    GetAclInformation := advapi32.NewProc("GetAclInformation")
    GetAce := advapi32.NewProc("GetAce")

    var sd, pp_dacl, pp_sacl windows.Handle
    var owner, group *windows.SID
    ret, _, _ := GetNamedSecurityInfo.Call(
        uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(path))),
        windows.SE_FILE_OBJECT,
        windows.DACL_SECURITY_INFORMATION,
        uintptr(unsafe.Pointer(&owner)),
        uintptr(unsafe.Pointer(&group)),
        uintptr(unsafe.Pointer(&pp_dacl)),
        uintptr(unsafe.Pointer(&pp_sacl)),
        uintptr(unsafe.Pointer(&sd)))
    if ret != 0 {
        log.Printf("GetNamedSecurityInfo(): %v", windows.Errno(ret))
        return
    }
    defer windows.LocalFree(sd)

    /***************************************************/
    var dacl windows.Handle
    var present, defaulted bool
    ret, _, _ = GetSecurityDescriptorDacl.Call(
        uintptr(unsafe.Pointer(&pp_dacl)),
        uintptr(unsafe.Pointer(&present)),
        uintptr(unsafe.Pointer(&dacl)),
        uintptr(unsafe.Pointer(&defaulted)))
    if ret != 0 {
        log.Printf("GetSecurityDescriptorDacl(): %v", windows.Errno(ret))
        return
    }
    log.Printf("GetSecurityDescriptorDacl(): %v", windows.Errno(ret))
    log.Printf("DACL Present/Defaulted: %t/%t", present, defaulted)
    log.Printf("DACL: %08x", dacl)
    /***************************************************/

    var aclSizeInfo AclSizeInformation
    ret, _, _ = GetAclInformation.Call(
        uintptr(unsafe.Pointer(&dacl)),
        uintptr(unsafe.Pointer(&aclSizeInfo)),
        unsafe.Sizeof(aclSizeInfo),
        uintptr(2))
    log.Printf("GetAclInformation(): %v", windows.Errno(ret))
    if ret != 0 {
        log.Printf("GetAclInformation(): %v", windows.Errno(ret))
        return
    }

    log.Printf("ACL SizeInfo: %08x", aclSizeInfo)
    /***************************************************/
    for i := uint32(0); i < aclSizeInfo.AceCount; i++ {
        var ace *AccessAllowedAce
        ret, _, _ = GetAce.Call(
            uintptr(unsafe.Pointer(&dacl)),
            uintptr(i),
            uintptr(unsafe.Pointer(&ace)),
            0,
        )
        if ret != 0 {
            log.Printf(": GetAce(): %v", windows.Errno(ret))
            continue
        }

        // Print information about the ACE
        fmt.Printf("ACE %d:\n", i)
        //fmt.Printf("  Type: %d\n", ace.Header.AceType)
        //fmt.Printf("  Flags: %d\n", ace.Header.AceFlags)
        //fmt.Printf("  Access Mask: %d\n", ace.Mask)
    }
}
