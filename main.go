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

func scan(rootPath *string) {
	for filePath, mostRecentModTime := range walk(rootPath) {
		if lastModTime, ok := watchedFiles[filePath]; ok {
			if lastModTime != mostRecentModTime {
				watchedFiles[filePath] = mostRecentModTime
				runOrSkipTest(filePath)
			}
		} else {
			// files recently added
			fileInfo, _ := os.Stat(filePath)
			watchedFiles[filePath] = fileInfo.ModTime()
		}
	}
}

func walk(rootPath *string) FileInfo {
	watchedFiles := FileInfo{}
	err := filepath.Walk(*rootPath, visit(watchedFiles))
	if err != nil {
		panic(err)
	}

	return watchedFiles
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
	return !strings.HasSuffix(fileName, "_test.go")
}

func test(filePath string) {
	clear()
	cmd := exec.Command("go", "test", "-cover", path.Dir(filePath))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Run()
	outStr := string(stdout.Bytes())
	fmt.Printf("%s", outStr)
}

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
