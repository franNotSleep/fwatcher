package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	filenames = flag.String("i", "", "The files to include")
	cmdStr    = flag.String("c", "", "The command to execute when change detected")
)

type FWatchFS struct {
	Files       map[string]*os.File
	lastModTime time.Time
}

type FWModInfo struct {
	HasChanged bool
	Info       *os.FileInfo
	File       *os.File
}

func NewFWatchFS() *FWatchFS {
	return &FWatchFS{Files: make(map[string]*os.File), lastModTime: time.Now()}
}

func (fw *FWatchFS) HasChanged(filename string) *FWModInfo {
	var fwModInfo FWModInfo

	f, ok := fw.Files[filename]
	if !ok {
		of, err := os.Open(filename)
		if err != nil {
			panic(err)
		}

		f = of
		fw.Files[filename] = of
	}

	info, err := f.Stat()
	if err != nil {
		panic(err)
	}

	fwModInfo.Info = &info
	fwModInfo.File = f

	if filename == "main.go" {
		fmt.Printf("info mod: %s\n", info.ModTime())
	} else {
		fmt.Printf("filename %s\n", filename)
	}

	if info.ModTime().UnixMilli() > fw.lastModTime.UnixMilli() {
		fw.lastModTime = info.ModTime()
		fwModInfo.HasChanged = true
		return &fwModInfo
	}

	fwModInfo.HasChanged = false
	return &fwModInfo
}

func hasChange(filenames []string, fw *FWatchFS) bool {
	for _, filename := range filenames {
		if filename == "" {
			continue
		}

		changeInfo := fw.HasChanged(filename)
		if changeInfo.HasChanged {
			return true
		}

		if (*changeInfo.Info).IsDir() {
			dirnames, err := (*changeInfo.File).ReadDir(0)
			fmt.Printf("%+v\n", dirnames)

			if err != nil {
				log.Fatal(err)
			}

			names := make([]string, len(dirnames))
			for _, dirname := range dirnames {
				fmt.Printf("name: %s\n", dirname.Name())
				names = append(names, fmt.Sprintf("%s/%s", filename, dirname.Name()))
			}

			if hasChange(names, fw) {
				return true
			}
		}
	}

	return false
}

func main() {
	flag.Parse()
	if *cmdStr == "" {
		fmt.Printf("command not specified")
		os.Exit(1)
	}

	cmdSplit := strings.Split((*cmdStr), " ")
	cmdName, cmdArgs := cmdSplit[0], cmdSplit[1:]

	watchFilenames := strings.Split(*filenames, ",")

	fw := NewFWatchFS()
	for {
		if hasChange(watchFilenames, fw) {
			command := exec.Command(cmdName, cmdArgs...)
			command.Stdout = os.Stdout

			if err := command.Run(); err != nil {
				log.Fatal(err)
			}
		}
	}
}
