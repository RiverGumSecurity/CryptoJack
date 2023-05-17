package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
    "sync"

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
	arg_ndir := flag.Int("n", 1, "how many directories to create")
	flag.Parse()
	startdir, _ := filepath.Abs(strings.Title(*arg_directory))

	fmt.Printf("[*] Starting directory is: [%s], max depth = %d.\n", startdir, *arg_fakedepth)
	if _, err := os.Stat(startdir); err != nil {
		err := os.Mkdir(startdir, 0755)
		if err != nil {
			panic(err)
		}
	}

    var wg sync.WaitGroup
    for i := 0; i < *arg_ndir; i++ {
        si := fmt.Sprintf("%02d", i)
        dir := filepath.Join(startdir, si)
		os.Mkdir(dir, 0755)
        wg.Add(1)
	    go cjlib.FakeData(dir, *arg_fakedepth, &wg)
    }
    wg.Wait()
}
