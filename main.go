package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// FileInfo ...
type FileInfo map[string]time.Time

var watchedFiles = FileInfo{}

func main() {
	defaultPath, _ := os.Getwd()
	rootPath := flag.String("path", defaultPath, "the root path to be watched")
	watchPeriod := flag.Duration("time", 10*time.Second, "the period (in seconds) for watching files")
	flag.Parse()

	watchedFiles = walk(rootPath)
	runTicker := time.NewTicker(*watchPeriod)
	for {
		select {
		case <-runTicker.C:
			watch(rootPath)
		}
	}
}

func watch(path *string) {
	newWatchedFiles := walk(path)
	for path, lastModified := range newWatchedFiles {
		if modifiedAt, ok := watchedFiles[path]; ok {
			fmt.Println("________________")
			fmt.Println(path)
			fmt.Printf("Modified at: %v\n", modifiedAt)
			fmt.Printf("Last modified: %v\n", lastModified)

			if modifiedAt != lastModified {
				fmt.Printf("Running tests for %s\n", path)
				watchedFiles[path] = lastModified
				runCmd(path)
			}
		}
	}
}

func walk(rootPath *string) FileInfo {
	newWatchedFiles := FileInfo{}
	err := filepath.Walk(*rootPath, visit(newWatchedFiles))
	if err != nil {
		panic(err)
	}

	return newWatchedFiles
}

func visit(watchedFiles FileInfo) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		if filepath.Ext(path) == ".go" {
			watchedFiles[path] = info.ModTime()
		}

		return nil
	}
}

func runCmd(path string) {
	cmd := exec.Command("go", "test", path)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
}
