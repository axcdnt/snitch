package main

import (
	"flag"
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
	changedDirs := make(map[string]bool, 0)
	for filePath, mostRecentModTime := range walk(rootPath) {
		lastModTime, ok := watchedFiles[filePath]
		if !ok {
			// files recently added
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				log.Print("Stat:", filePath, err)
				continue
			}
			watchedFiles[filePath] = fileInfo.ModTime()
			continue
		}

		// no changes
		if lastModTime == mostRecentModTime {
			continue
		}

		watchedFiles[filePath] = mostRecentModTime
		if !needsTest(filePath) {
			continue
		}

		pkgDir := path.Dir(filePath)
		changedDirs[pkgDir] = true
	}

	for dir := range changedDirs {
		test(dir)
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

// needsTest returns if we need to run test for the changed path.
func needsTest(filePath string) bool {
	if isTestFile(filePath) {
		// change happened on a test file, we need to test.
		return true
	}

	// changed file is not a test file, we return true only if there is a
	// test file for the go file.
	fileName := path.Base(filePath)
	path := path.Dir(filePath)
	extensionLessFileName := strings.Split(fileName, ".")[0]
	testFilePath := filepath.Join(path, extensionLessFileName+"_test.go")
	_, ok := watchedFiles[testFilePath]
	return ok
}

func isTestFile(fileName string) bool {
	return strings.HasSuffix(fileName, "_test.go")
}

func test(dir string) {
	clear()
	cmd := exec.Command("go", "test", "-v", "-cover", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
