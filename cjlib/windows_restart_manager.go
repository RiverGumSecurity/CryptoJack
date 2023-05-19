package cjlib

import (
    "fmt"
    "math/rand"
    "syscall"
    "unsafe"
    ps "github.com/mitchellh/go-ps"
)

var targetProcesses = []string {
    "winword.exe",
    "excel.exe",
    "msedge.exe",
    "chrome.exe",
}

const (
    RmSessionKeyLen = 32
    CchRmMaxAppName = 255
    CchRmMaxSvcName = 63
    RmForceShutdown = 0x00000001
    RmShutdownOnlyRegistered = 0x00000010
    RmShutdownOnlyRestart = 0x00000020
)

type RmUniqueProcess struct {
    ProcessId uint32
    ProcessStartTime syscall.Filetime
}

type Proc struct {
    pid int
    ppid int
    name string
}

var (
    rsmdll = syscall.NewLazyDLL("rstrtmgr.dll")
    rmStartSession = rsmdll.NewProc("RmStartSession")
    rmRegisterResources = rsmdll.NewProc("RmRegisterResources")
    rmShutdown = rsmdll.NewProc("RmShutdown")
    rmEndSession = rsmdll.NewProc("RmEndSession")
)

func randString(n int) string {
    charset := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
    b := make([]byte, n)
    for i := range b {
        b[i] = charset[rand.Intn(len(charset))]
    }
    return string(b)
}

func rsmStartSession() (uint32, string, error) {
    var sessionHandle uint32 = 0
    sessionKey := randString(32)

    ret, _, _ := rmStartSession.Call(
        uintptr(unsafe.Pointer(&sessionHandle)), uintptr(0),
        uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(sessionKey))))
    if ret != 0 {
        return 0, "", fmt.Errorf("ERROR: RmStartSession(): %d", ret)
    }
    return sessionHandle, sessionKey, nil
}

func rsmRegisterProcesses(sessionHandle uint32, pids []int) error {
    if len(pids) == 0 {
        return fmt.Errorf("Empty process ID list")
    }

    uniqueProcesses := make([]RmUniqueProcess, len(pids))
    for i, pid := range pids {
        uniqueProcesses[i].ProcessId = uint32(pid)
    }
    ret, _, _ := rmRegisterResources.Call(
        uintptr(sessionHandle),
        0, 0, uintptr(len(uniqueProcesses)),
        uintptr(unsafe.Pointer(&uniqueProcesses[0])), 0, 0)
    if ret != 0 {
        return fmt.Errorf("failed to register processes. Error: %d", ret)
    }
    return nil
}

func rsmShutdown(sessionHandle uint32, shutdownFlags uint32) error {
    ret, _, _ := rmShutdown.Call(
        uintptr(sessionHandle), uintptr(shutdownFlags), 0)
    if ret != 0 {
        return fmt.Errorf("failed to initiate process shutdown. Error: %d", ret)
    }
    return nil
}

func rsmEndSession(sessionHandle uint32) error {
    ret, _, _ := rmEndSession.Call(uintptr(sessionHandle))
    if ret != 0 {
        return fmt.Errorf("failed to end the restart manager session. Error: %d", ret)
    }
    return nil
}

func inTargetProcs(name string) bool {
    for _, p := range targetProcesses {
        if name == p { return true }
    }
    return false
}

func isTargetProc(pid int, name string, procs []Proc) bool {
    n := 0
    found_ppid := false
    for _, p := range procs {
        if name != p.name || !inTargetProcs(p.name) { continue }
        if pid == p.ppid {
            found_ppid = true
        }
        n += 1
    }
    if (n > 1 && found_ppid) || n == 1 {
        return true
    }
    return false
}

func createPidList() ([]int, error) {
    processList, err := ps.Processes()
    if err != nil {
        return nil, err
    }
    // create a process snapshot first
    allProcs := make([]Proc, len(processList))
    for i, process := range processList {
        allProcs[i].pid = process.Pid()
        allProcs[i].ppid = process.PPid()
        allProcs[i].name = process.Executable()
    }

    // now lets find what we want
    var pidlist []int
    for _, p := range allProcs {
        if isTargetProc(p.pid, p.name, allProcs) {
            pidlist = append(pidlist, p.pid)
        }
    }
    return pidlist, nil
}

func Win32_RSMShutdownTargets() {
    sessionHandle, _, err := rsmStartSession()
    if err != nil {
        fmt.Println(err)
        return
    }
    defer rsmEndSession(sessionHandle)

    pidlist, _ := createPidList()
    fmt.Println(pidlist)

    err = rsmRegisterProcesses(sessionHandle, pidlist)
    if err != nil {
        fmt.Println(err)
        return
    }
    err = rsmShutdown(sessionHandle, RmShutdownOnlyRestart)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("[*] Win32_RSMShutdownTargets(): Completed Successfully.")
}


