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
    DECRYPTOR

    Version 1.0.1 by Joff Thyer
    Black Hills Information Security
    Copyright (c) 2022
__________________________________________________
`

	fmt.Println(banner)
	arg_dryrun := flag.Bool("n", false, "perform a dryrun without any decryption actions")
	arg_cryptext := flag.String("ext", ".cryptojack", "file extension to use for renamed content")
	arg_directory := flag.String("d", "", "Specify a starting directory. Default is $HOME directory.")
	arg_norename := flag.Bool("norename", false, "Dont rename files from encrypted back to original")
    arg_yaml := flag.String("y", "", "Specify a YAML IOC profile file name.")
	flag.Parse()

    if len(*arg_directory) == 0 {
        panic("You must specify the -d <directory> option")
    }
	if *arg_dryrun {
		fmt.Println("[*] DRY RUN Mode: No files will be decrypted.")
	} else {
		fmt.Printf(`
[*] --<[ WARNING ]>--    --<[ WARNING ]>--    --<[ WARNING ]>--
[*]
[*] You are about to decrypt ALL files recursively in the target
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

	_, _, _, err = cjlib.DecryptDirectoryStructure(*arg_directory, file_ext, *arg_norename, *arg_dryrun)
	if err != nil {
		fmt.Println(err.Error())
	}
}
