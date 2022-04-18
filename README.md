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
(*ioc_type*) are **url**, and **command**.

