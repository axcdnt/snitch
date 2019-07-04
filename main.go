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
	defaultPath, err := os.Getwd()
	if err != nil {
		log.Fatal("Getwd: ", err)
	}

	rootPath := flag.String("path", defaultPath, "the root path to be watched")
	interval := flag.Duration("interval", 5*time.Second, "the interval (in seconds) for scanning files")
	flag.Parse()

	if *interval < 0 {
		log.Fatal("negative interval:", *interval)
	}

	watchedFiles = walk(rootPath)
	for range time.NewTicker(*interval).C {
		scan(rootPath)
	}
}

func scan(rootPath *string) {
	for filePath, mostRecentModTime := range walk(rootPath) {
		lastModTime, ok := watchedFiles[filePath]
		if !ok {
			// files recently added
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				log.Print("Stat:", filePath, err)
				return
			}
			watchedFiles[filePath] = fileInfo.ModTime()
			return
		}

		// no changes
		if lastModTime == mostRecentModTime {
			return
		}

		watchedFiles[filePath] = mostRecentModTime
		runOrSkipTest(filePath)
	}
}

func walk(rootPath *string) FileInfo {
	wf := FileInfo{}
	if err := filepath.Walk(*rootPath, visit(wf)); err != nil {
		log.Fatal("walk:", err)
	}
	return wf
}

func visit(wf FileInfo) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".go" {
			wf[path] = info.ModTime()
		}

		return nil
	}
}

func runOrSkipTest(filePath string) {
	testFilePath := findTestFilePath(filePath)
	if testFilePath == "" {
		// there is no test file for the modified path, ignore.
		return
	}
	test(testFilePath)
}

func findTestFilePath(filePath string) string {
	if isTestFile(filePath) {
		// change happened on a test file, just return its path.
		return filePath
	}

	// changed file is not a test file, we need to change its path to
	// contain an _test.go suffix then return the full path. if test
	// file for the changed path does not exist, returns empty string
	// instead.
	fileName := path.Base(filePath)
	path := path.Dir(filePath)
	extensionLessFileName := strings.Split(fileName, ".")[0]
	testFilePath := filepath.Join(path, extensionLessFileName+"_test.go")
	if _, ok := watchedFiles[testFilePath]; !ok {
		// no test file exists, return empty.
		return ""
	}

	return testFilePath
}

func isTestFile(fileName string) bool {
	return strings.HasSuffix(fileName, "_test.go")
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
