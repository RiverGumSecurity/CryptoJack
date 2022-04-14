package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
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
    FAKE DATA
	
    Version 1.0.1 by Joff Thyer
    Black Hills Information Security
    Copyright (c) 2022
__________________________________________________
`

	fmt.Println(banner)
	arg_directory := flag.String("d", strings.Title(cjlib.RandomWord()), "Specify a starting directory.")
	arg_fakedepth := flag.Int("depth", 2, "how deep to recurse when creating fake data structure")
	flag.Parse()
	startdir, _ := filepath.Abs(strings.Title(*arg_directory))

	fmt.Printf("[*] Fake data directory is: [%s], max depth = %d.\n", startdir, *arg_fakedepth)
	fmt.Printf("[*] DO YOU WANT TO PROCEED [Y|N]? ")
	ans := []byte("N")
	os.Stdin.Read(ans)
	if ans[0] != 89 {
		os.Exit(0)
	}

	// panic if starting dir is not real!
	if _, err := os.Stat(startdir); err != nil {
		err := os.Mkdir(startdir, 0755)
		if err != nil {
			panic(err)
		}
	}

	msg := make(chan string)
	go cjlib.FakeData(startdir, *arg_fakedepth, msg)
	result := <-msg
	fmt.Println(result)
}
