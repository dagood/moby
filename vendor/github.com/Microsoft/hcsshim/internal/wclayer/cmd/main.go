package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/Microsoft/go-winio"
	"github.com/Microsoft/hcsshim/internal/safefile"
	"github.com/Microsoft/hcsshim/internal/winapi"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	// Mock setup
	os.WriteFile("tempThing.txt", []byte("hello"), 0o666)
	os.Mkdir("rootDir", 0o777)

	for i := 0; i < 100000; i++ {
		if err := test(); err != nil {
			return err
		}
	}
	return nil
}

var n int

func test() error {
	// Mock setup
	fname := "foo.txt"
	wroot, _ := os.OpenFile("rootDir", os.O_RDONLY, 0)
	tempf, _ := os.Open("tempThing.txt")
	fileInfo, _ := winio.GetFileBasicInfo(tempf)

	n++
	// From legacy.go
	f, err := safefile.OpenRelative(fname, wroot, syscall.GENERIC_READ|syscall.GENERIC_WRITE, syscall.FILE_SHARE_READ, winapi.FILE_CREATE, 0)
	if err != nil {
		// Fails after 3151, 788, 2353, ... random number of iterations, whether in Go 1.21.5 or 1.22rc1.
		// > openRelative: 1310, open rootDir\foo.txt: Access is denied.
		return fmt.Errorf("openRelative: %v, %v", n, err)
	}
	defer func() {
		if f != nil {
			f.Close()
			_ = safefile.RemoveRelative(fname, wroot)
		}
	}()

	strippedFi := *fileInfo
	strippedFi.FileAttributes = 0
	err = winio.SetFileBasicInfo(f, &strippedFi)
	if err != nil {
		return fmt.Errorf("set info: %v", err)
	}

	//...
	return nil
}
