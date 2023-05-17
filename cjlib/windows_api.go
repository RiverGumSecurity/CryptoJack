package cjlib

import (
    "fmt"
    "unsafe"
    "golang.org/x/sys/windows"
)

type systemInfo struct {
    wProcessorArchitecture      uint16
    wReserved                   uint16
    dwPageSize                  uint32
    lpMinimumApplicationAddress uintptr
    lpMaximumApplicationAddress uintptr
    dwActiveProcessorMask       uintptr
    dwNumberOfProcessors        uint32
    dwProcessorType             uint32
    dwAllocationGranularity     uint32
    wProcessorLevel             uint16
    wProcessorRevision          uint16
}

func bitsToDrives(bitMap uint32) (drives []string) {
    availableDrives := []string{
        "A", "B", "C", "D", "E", "F", "G", "H",
        "I", "J", "K", "L", "M", "N", "O", "P",
        "Q", "R", "S", "T", "U", "V", "W", "X",
        "Y", "Z"}
    for i := range availableDrives {
        if bitMap & 1 == 1 {
            drives = append(drives, availableDrives[i] + ":")
        }
        bitMap >>= 1
    }
    return
}

func Win32_GetLogicalDrives() {
    kernel32 := windows.NewLazySystemDLL("kernel32.dll")
    GetLogicalDrives := kernel32.NewProc("GetLogicalDrives")
    var drives []string
    ret, _, err := GetLogicalDrives.Call()
    if err.Error() == "The operation completed successfully." {
        drives = bitsToDrives(uint32(ret))
        fmt.Printf("[+] GetLogicalDrives(): %v\n", drives)
    } else {
        fmt.Printf("[-] %s\n", err.Error())
    }
}

func Win32_GetNativeSystemInfo() {
    kernel32 := windows.NewLazySystemDLL("kernel32.dll")
    GetNativeSystemInfo := kernel32.NewProc("GetNativeSystemInfo")
    var sysinfo systemInfo
    _, _, err := GetNativeSystemInfo.Call(uintptr(unsafe.Pointer(&sysinfo)))
    if err.Error() == "The operation completed successfully." {
        fmt.Printf(`[*] GetNativeSystemInfo():
    [+] Processor Architecture : 0x%08X
    [+] Page Size              : %d
    [+] Number of CPU's        : %d
    [+] Processor Level        : 0x%08X
    [+] Processor Revision     : 0x%08X
`,      sysinfo.wProcessorArchitecture, sysinfo.dwPageSize, sysinfo.dwNumberOfProcessors,
        sysinfo.wProcessorLevel, sysinfo.wProcessorRevision)
    } else {
        fmt.Printf("[-] %s\n", err.Error())
    }
}
