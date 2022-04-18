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
