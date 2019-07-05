package main

import (
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
		log.Fatal("invalid interval, must be > 0:", *interval)
	}

	if err := os.Chdir(*rootPath); err != nil {
		log.Fatal("could not change directory:", err)
	}

	watchedFiles = walk(rootPath)
	for range time.NewTicker(*interval).C {
		scan(rootPath)
	}
}

func scan(rootPath *string) {
	modifiedDirs := make(map[string]bool, 0)
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
		modifiedDirs[pkgDir] = true
	}

	if len(modifiedDirs) == 0 {
		return
	}

	dedup := make([]string, 0)
	for dir := range modifiedDirs {
		dedup = append(dedup, dir)
	}
	test(dedup)
}

func walk(rootPath *string) FileInfo {
	wf := FileInfo{}
	if err := filepath.Walk(*rootPath, visit(wf)); err != nil {
		log.Fatal("could not traverse files:", err)
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
	return isTestFile(filePath) || hasTestFile(filePath)
}

func isTestFile(fileName string) bool {
	return strings.HasSuffix(fileName, "_test.go")
}

// returns true if a go test file exists for filePath.
func hasTestFile(filePath string) bool {
	ext := filepath.Ext(filePath)
	testFilePath := fmt.Sprintf(
		"%s_test.go",
		filePath[0:len(filePath)-len(ext)],
	)
	_, ok := watchedFiles[testFilePath]
	return ok
}

func test(dirs []string) {
	clear()
	for _, dir := range dirs {
		cmd := exec.Command("go", "test", "-v", "-cover", dir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
}

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
