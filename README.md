# CryptoJack

CryptoJack is a ransomware simulation program which can be used to check whether current
defenses are able to detect ransomware activity.  CryptoJack has no built in exploitation
or spreading ability but rather focuses on the core activity of recursively encrypting
files in a specified directory. CryptoJack is written in GOlang and will execute in either
a Linux or Windows environment.

*WARNING: When using this tool, DO NOT DELETE THE ENCRYPTION KEY FILE(S). Any encrypted data will NOT BE
RECOVERABLE if you DELETE the encryption key file(s).*

CryptoJack consists of four components:

* fakedata: the ability to generate a recursive directory structure with fake documents in each directory.
* encrypt: the encryption program itself.
* decrypt: the decryption program.
* rbot: an experimental Discord bot that allows for remote command operation.

## Typical Usage Sequence

Most people will be comforted by the idea that you can create a fake data directory
structure so that using this ransomware threat emulation tool does not impact production data.
The **fakedata** tool with default options will create a fake data structure with a randomly
chosen english word as the starting directory. You can also specify this if you desire.

```
PS C:\demo> .\fakedata.exe -depth 1

_________________________________________________

    ╔═╗┬─┐┬ ┬┌─┐┌┬┐┌─┐ ╦┌─┐┌─┐┬┌─
    ║  ├┬┘└┬┘├─┘ │ │ │ ║├─┤│  ├┴┐
    ╚═╝┴└─ ┴ ┴   ┴ └─┘╚╝┴ ┴└─┘┴ ┴
    FAKE DATA

    Version 1.0.1 by Joff Thyer
    Black Hills Information Security
    Copyright (c) 2022
__________________________________________________

[*] Fake data directory is: [C:\demo\Bosom], max depth = 1.
[*] DO YOU WANT TO PROCEED [Y|N]? Y
2022/04/18 10:45:08 Fake Data Creation: [C:\demo\Bosom]
2022/04/18 10:45:08 Fake Data Creation: [C:\demo\Bosom\Partner]
2022/04/18 10:45:08 Fake Data Creation: [C:\demo\Bosom\Wage]
2022/04/18 10:45:08 Fake Data Creation: [C:\demo\Bosom\Continent]
2022/04/18 10:45:08 Fake Data Creation: [C:\demo\Bosom\Psychologist]
2022/04/18 10:45:08 Fake Data Creation: [C:\demo\Bosom\Allocation]
Created 5 fake directories and 70 files in [C:\demo\Bosom]!
PS C:\demo> .\fakedata.exe
```

After creating fake data, you would then execute the **encrypt** operation providing a YML
IOC profile using the "-y" argument if you so desire. When encryption is complete, it will
write the encryption key files, database hashing file, and a file called *__RansomNote__.html*
as well as opening up the ransom note in a browser.


```
PS C:\demo> .\encrypt.exe -d .\Bosom\

_________________________________________________

    ╔═╗┬─┐┬ ┬┌─┐┌┬┐┌─┐ ╦┌─┐┌─┐┬┌─
    ║  ├┬┘└┬┘├─┘ │ │ │ ║├─┤│  ├┴┐
    ╚═╝┴└─ ┴ ┴   ┴ └─┘╚╝┴ ┴└─┘┴ ┴
    ENCRYPTOR

    Version 1.0.1 by Joff Thyer
    Black Hills Information Security
    Copyright (c) 2022
__________________________________________________


[*] --<[ WARNING ]>--    --<[ WARNING ]>--    --<[ WARNING ]>--
[*]
[*] You are about to encrypt ALL files recursively in the target
[*] directory: [.\Bosom\]
[*]
[*] --<[ WARNING ]>--    --<[ WARNING ]>--    --<[ WARNING ]>--

[*] DO YOU REALLY WANT TO PROCEED [Y|N]? Y
2022/04/18 10:46:23 ENCRYPTED 1/70: C:\demo\Bosom\AlcoveAtrium.pdf
2022/04/18 10:46:24 ENCRYPTED 2/70: C:\demo\Bosom\AmnestyTownhouse.xlsx
2022/04/18 10:46:24 ENCRYPTED 3/70: C:\demo\Bosom\AnchovyInfarction.xlsx
2022/04/18 10:46:24 ENCRYPTED 4/70: C:\demo\Bosom\BiplaneCounterterrorism.pdf
2022/04/18 10:46:24 ENCRYPTED 5/70: C:\demo\Bosom\BoughSeafood.xlsx
2022/04/18 10:46:24 ENCRYPTED 6/70: C:\demo\Bosom\Continent\AcademyAirforce.xlsx
2022/04/18 10:46:24 ENCRYPTED 7/70: C:\demo\Bosom\Continent\ApparelUnemployment.pdf
```


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
