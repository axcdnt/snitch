package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// FileInfo ...
type FileInfo map[string]time.Time

var watchedFiles = FileInfo{}

func main() {
	defaultPath, _ := os.Getwd()
	rootPath := flag.String("path", defaultPath, "the root path to be watched")
	interval := flag.Int("interval", 5, "the interval (in seconds) for scanning files")
	flag.Parse()

	watchedFiles = walk(rootPath)
	scheduleScanAt(rootPath, interval)
}

func scheduleScanAt(rootPath *string, interval *int) {
	runTicker := time.NewTicker(
		time.Duration(time.Duration(*interval) * time.Second))
	for {
		select {
		case <-runTicker.C:
			scan(rootPath)
		}
	}
}

func scan(basePath *string) {
	for filePath, mostRecentModTime := range walk(basePath) {
		if lastModTime, ok := watchedFiles[filePath]; ok {
			if lastModTime != mostRecentModTime {
				watchedFiles[filePath] = mostRecentModTime
				runOrSkipTest(filePath)
			}
		}
	}
}

func runOrSkipTest(filePath string) {
	if isRegularFile(filePath) {
		// run the respective test for go files
		testFilePath := findTestFilePath(filePath)
		if testFilePath != "" {
			test(testFilePath)
		}
	} else {
		// run tests directly
		test(filePath)
	}
}

func walk(rootPath *string) FileInfo {
	reWatchedFiles := FileInfo{}
	err := filepath.Walk(*rootPath, visit(reWatchedFiles))
	if err != nil {
		panic(err)
	}

	return reWatchedFiles
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

func test(filePath string) {
	cmd := exec.Command("go", "test", "-run", filePath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Run()
	outStr := string(stdout.Bytes())
	fmt.Printf("%s", outStr)
}

func findTestFilePath(filePath string) string {
	fileName := path.Base(filePath)
	if isRegularFile(fileName) {
		path := path.Dir(filePath)
		extensionLessFileName := strings.Split(fileName, ".")[0]
		testFilePath := filepath.Join(path, extensionLessFileName+"_test.go")
		if _, ok := watchedFiles[testFilePath]; ok {
			return testFilePath
		}
	}

	return ""
}

func isRegularFile(fileName string) bool {
	if !strings.HasSuffix(fileName, "_test.go") {
		return true
	}

	return false
}
