package cjlib

import (
    "fmt"
    "math/rand"
    "syscall"
    "unsafe"
    ps "github.com/mitchellh/go-ps"
)

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

var (
    rsmdll = syscall.NewLazyDLL("rstrtmgr.dll")
    rmStartSession = rsmdll.NewProc("RmStartSession")
    rmRegisterResources = rsmdll.NewProc("RmRegisterResources")
    rmShutdown = rsmdll.NewProc("RmShutdown")
    rmEndSession = rsmdll.NewProc("RmEndSession")
)

var targetProcesses = []string { "chrome.exe" }
//var targetProcesses = []string { "notepad.exe" }

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

func rsmRegisterProcesses(sessionHandle uint32, pids []uint32) error {
    if len(pids) == 0 {
        return fmt.Errorf("Empty process ID list")
    }

    uniqueProcesses := make([]RmUniqueProcess, len(pids))
    for i, pid := range pids {
        uniqueProcesses[i].ProcessId = pid
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

func inSlice(targetSlice []string, s string) bool {
    for i, _ := range targetSlice {
        if s == targetSlice[i] { return true }
    }
    return false
}

func createPidList() ([]uint32, error) {
    var pidlist []uint32
    processList, err := ps.Processes()
    if err != nil {
        return pidlist, err
    }
    for x := range processList {
        var process ps.Process
        process = processList[x]
        if inSlice(targetProcesses, process.Executable()) {
            pidlist = append(pidlist, uint32(process.Pid()))
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
    err = rsmRegisterProcesses(sessionHandle, pidlist)
    if err != nil {
        fmt.Println(err)
        return
    }
    err = rsmShutdown(sessionHandle, RmShutdownOnlyRegistered)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("[*] Process shutdown initiated successfully.")
}


