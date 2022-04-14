package cjlib

import (
    "net"
    "strings"
)

func inslice(s []string, str string) bool {
    for _, v := range s {
        if v == str {
            return true
        }
    }
    return false
}

func AddressList() ([]string, error) {
    var retlist []string
    ifaces, err := net.Interfaces()
    if err != nil {
        return retlist, err
    }

    for _, i := range ifaces {
        addrs, err := i.Addrs()
        if err != nil ||
                    !strings.Contains(i.Flags.String(), "up") ||
                    strings.HasPrefix(strings.ToLower(i.Name), "lo") ||
                    strings.HasPrefix(strings.ToLower(i.Name), "vether") {
            continue
        }
        for _, a := range addrs {
            if !inslice(retlist, a.String()) {
                retlist = append(retlist, a.String())
            }
        }
    }
    return retlist, nil
}
