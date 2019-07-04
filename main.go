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
	scheduleScanAt(rootPath, interval)
}

func scheduleScanAt(rootPath *string, interval *time.Duration) {
	runTicker := time.NewTicker(*interval)
	for {
		select {
		case <-runTicker.C:
			scan(rootPath)
		}
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
	watchedFiles := FileInfo{}
	if err := filepath.Walk(*rootPath, visit(watchedFiles)); err != nil {
		log.Fatal("walk:", err)
	}

	return watchedFiles
}

func visit(watchedFiles FileInfo) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
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
