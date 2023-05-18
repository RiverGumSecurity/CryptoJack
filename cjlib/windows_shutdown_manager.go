package cjlib

import (
    "fmt"
    "golang.org/x/sys/windows"
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
    ProcessId         uint32
    ProcessStartTime syscall.Filetime
}

var (
    rsmdll = windows.NewLazyDLL("rstrtmgr.dll")
    rmStartSession = rsmdll.NewProc("RmStartSession")
    rmRegisterResources = rsmdll.NewProc("RmRegisterResources")
    rmShutdown = rsmdll.NewProc("RmShutdown")
    rmEndSession = rsmdll.NewProc("RmEndSession")
)

var targetProcesses = []string { "msedge.exe", "chrome.exe" }

func rsmStartSession() (uintptr, string, error) {
    var sessionHandle uintptr
    var sessionKey [RmSessionKeyLen]uint16
    ret, _, err := rmStartSession.Call(
        sessionHandle, 0, uintptr(unsafe.Pointer(&sessionKey[0])))
    fmt.Printf("rmStartSession(): %d\n", ret)
    if err.Error() != "The operation completed successfully." {
        return 0, "", err
    }
    return sessionHandle, syscall.UTF16ToString(sessionKey[:]), nil
}

func rsmRegisterProcesses(sessionHandle uintptr, processIds []uint32) error {
    if len(processIds) == 0 { return nil }
    uniqueProcesses := make([]RmUniqueProcess, len(processIds))
    for i, pid := range processIds {
        uniqueProcesses[i].ProcessId = pid
    }
    ret, _, err := rmRegisterResources.Call(
        sessionHandle, 0, 0, uintptr(len(uniqueProcesses)),
        uintptr(unsafe.Pointer(&uniqueProcesses[0])), 0, 0)
    fmt.Printf("rmRegisterResources(): %d\n", ret)
    if err.Error() != "The operation completed successfully." {
        return err
    }
    return nil
}

func rsmShutdownProcesses(sessionHandle uintptr, shutdownFlags uint32, restartType uint32) error {
    ret, _, err := rmShutdown.Call(
        sessionHandle, uintptr(shutdownFlags), uintptr(restartType), 0)
    fmt.Printf("rmShutdown(): %d\n", ret)
    if err.Error() != "The operation completed successfully." {
        return err
    }
    return nil
}

func rsmEndSession(sessionHandle uintptr) error {
    _, _, err := rmEndSession.Call(sessionHandle)
    if err.Error() != "The operation completed successfully." {
        return err
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
    err = rsmShutdownProcesses(sessionHandle, 0, RmShutdownOnlyRestart)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("[*] Process shutdown initiated successfully.")
}


