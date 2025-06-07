package main

import (
	"errors"
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
}

func NewFWatchFS() *FWatchFS {
	return &FWatchFS{Files: make(map[string]*os.File), lastModTime: time.Now()}
}

func (fw *FWatchFS) HasChanged(filename string) (*FWModInfo, error) {
	var fwModInfo FWModInfo

	info, err := os.Stat(filename)
	if err != nil {
		return &fwModInfo, err
	}

	fwModInfo.Info = &info

	if info.ModTime().UnixMilli() > fw.lastModTime.UnixMilli() {
		fw.lastModTime = info.ModTime()
		fwModInfo.HasChanged = true
		return &fwModInfo, nil
	}

	fwModInfo.HasChanged = false
	return &fwModInfo, nil
}

func hasChange(filenames []string, fw *FWatchFS) bool {
	for _, filename := range filenames {
		if filename == "" {
			continue
		}

		changeInfo, err := fw.HasChanged(filename)
		if err != nil {

			if errors.Is(err, os.ErrNotExist) {
				// this may be a symlic pointing to an empty location
				return false
			}
			panic(err)
		}

		if changeInfo.HasChanged {
			return true
		}

		if (*changeInfo.Info).IsDir() {
			dirnames, err := os.ReadDir(filename)

			if err != nil {
				log.Fatal(err)
			}

			names := make([]string, len(dirnames))
			for _, dirname := range dirnames {
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
