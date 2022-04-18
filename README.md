# CryptoJack

CryptoJack is a ransomware simulation program which can be used to check whether current
defenses are able to detect ransomware activity.  CryptoJack has no built in exploitation
or spreading ability but rather focuses on the core activity of recursively encrypting
files in a specified directory. CryptoJack is written in GOlang and will execute in either
a Linux or Windows environment.

CryptoJack consists of four components:

* fakedata: the ability to generate a recursive directory structure with fake documents in each directory.
* encrypt: the encryption program itself.
* decrypt: the decryption program.
* rbot: an experimental Discord bot that allows for remote command operation.

## YML IOC Profiles

Within the CryptoJack distribution is a *yml* directory containing profiles
used for threat emulation. As of today, the only two IOC types supported
(*ioc_type*) are **url**, and **command**. In the case of **url**, the code will
attempt an HTTP GET request during *encrypt* for any URL specified.
In the latter case of **command**, an *echo* statement is always placed in front of any executed O/S command.


### Example YML Profile

```
- data: https://docs.google.com/spreadsheets/d/11C7pdR3r_VeOPQXpRCGtUEJoftKO1wB7ZFfX0t94XTw/edit#gid=0&range=B1
  ioc_type: url
  note: LockBit backdoor installer loader URL
- data: vssadmin Delete Shadows /All /Quiet
  ioc_type: command
  note: LockBit
- data: vssadmin delete shadows /all /quiet & wmic shadowcopy delete & bcdedit /set {default} bootstatuspolicy ignoreallfailures & bcdedit /set {default} recoveryenabled no & wbadmin delete catalog -quiet
  ioc_type: command
  note: LockBit
```

### Usage: Encrypt

```

_________________________________________________

    ╔═╗┬─┐┬ ┬┌─┐┌┬┐┌─┐ ╦┌─┐┌─┐┬┌─
    ║  ├┬┘└┬┘├─┘ │ │ │ ║├─┤│  ├┴┐
    ╚═╝┴└─ ┴ ┴   ┴ └─┘╚╝┴ ┴└─┘┴ ┴
    ENCRYPTOR

    Version 1.0.1 by Joff Thyer
    Black Hills Information Security
    Copyright (c) 2022
__________________________________________________

Usage of C:\Users\jsthyer\Projects\CryptoJack\encrypt.exe:
  -d string
        Specify a starting directory. This is required.
  -e string
        filename extensions which will be excluded (default "exe, dll")
  -ext string
        file extension to use for renamed content (default ".cryptojack")
  -n    perform a dryrun without any encryption actions
  -norename
        Dont rename files to encrypted filename + extension
  -y string
        Specify a YAML IOC profile file name.
```

