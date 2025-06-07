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

func hasChange(filenames []string, lastModTime *time.Time) bool {
	for _, filename := range filenames {
		if filename == "" {
			continue
		}
		file, err := os.Open(filename)
		if err != nil {
			panic(err)
		}

		info, err := file.Stat()
		if err != nil {
			panic(err)
		}

		if info.ModTime().UnixMilli() > lastModTime.UnixMilli() {
			*lastModTime = info.ModTime()
			return true
		}

		if info.IsDir() {
			dirnames, err := file.ReadDir(0)

			if err != nil {
				log.Fatal(err)
			}

			names := make([]string, len(dirnames))
			for _, dirname := range dirnames {
				names = append(names, fmt.Sprintf("%s/%s", filename, dirname.Name()))
			}

			if hasChange(names, lastModTime) {
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
	lastModTime := time.Now()

	for {
		if hasChange(watchFilenames, &lastModTime) {
			command := exec.Command(cmdName, cmdArgs...)
			command.Stdout = os.Stdout

			if err := command.Run(); err != nil {
				log.Fatal(err)
			}
		}
	}
}
