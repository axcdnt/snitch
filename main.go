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
	watchInterval := flag.Duration("interval", 5*time.Second, "the interval (in seconds) for scanning files")
	flag.Parse()

	watchedFiles = walk(rootPath)
	runTicker := time.NewTicker(*watchInterval)
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
	// for a go file, run its respective test
	if isRegularFile(filePath) {
		testFilePath := findTestFilePath(filePath)
		if testFilePath != "" {
			test(testFilePath)
		}
	} else {
		// run _test.go files
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

func test(path string) {
	cmd := exec.Command("go", "test", "-run", path)
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
		fileWithoutExtension := strings.Split(fileName, ".")[0]
		testFilePath := filepath.Join(path, fileWithoutExtension+"_test.go")
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
