package cjlib

import (
    "syscall"
    "unsafe"
)

const (
    NetSetupUnknownStatus = iota
    NetSetupUnjoined
    NetSetupWorkgroupName
    NetSetupDomainName
)

var (
    netapi32 = syscall.NewLazyDLL("netapi32.dll")
    netGetJoinInformation = netapi32.NewProc("NetGetJoinInformation")
    secur32 = syscall.NewLazyDLL("secur32.dll")
    getComputerObjectName = secur32.NewProc("GetComputerObjectNameW")
)

func WindowsDomainStatus(computerName string) string {
    var joinStatus uint32
    _, _, _ = netGetJoinInformation.Call(
        uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(computerName))),
        uintptr(unsafe.Pointer(&joinStatus)),
    )
    switch joinStatus {
        case NetSetupUnknownStatus:
            return "Unknown"
        case NetSetupUnjoined:
            return "Not Domain Joined"
        case NetSetupWorkgroupName:
            return "Workgroup Joined"
        case NetSetupDomainName:
            return "Domain Joined"
        default:
            return "Unknown"
    }
}

func WindowsComputerName() string {
    var domainName [256]uint16
    var domainNameLen uint32 = uint32(len(domainName))
    _, _, _ = getComputerObjectName.Call(
        uintptr(2),
        uintptr(unsafe.Pointer(&domainName[0])),
        uintptr(unsafe.Pointer(&domainNameLen)),
    )
    return syscall.UTF16ToString(domainName[:]) 
}
