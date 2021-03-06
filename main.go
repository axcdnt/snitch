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

	"github.com/axcdnt/snitch/platform"
	"github.com/fatih/color"
)

// FileInfo represents a file and its modification date
type FileInfo map[string]time.Time

var (
	notifier platform.Notifier
	version  string
	pass     = color.New(color.FgGreen)
	fail     = color.New(color.FgHiRed)
)

func init() {
	notifier = platform.NewNotifier()
}

func main() {
	defaultPath, err := os.Getwd()
	if err != nil {
		log.Fatal("could not get current directory: ", err)
	}

	versionFlag := flag.Bool("v", false, "prints the current version and exit")
	rootPath := flag.String("path", defaultPath, "defines the root path to be watched")
	interval := flag.Duration("interval", 1*time.Second, "defines the interval (in seconds) for scanning files")
	flag.Parse()

	if *versionFlag {
		log.Printf("Build %s", version)
		return
	}

	if *interval < 0 {
		log.Fatal("invalid interval, must be > 0", *interval)
	}

	if err := os.Chdir(*rootPath); err != nil {
		log.Fatal("could not change directory: ", err)
	}

	log.Print("Snitch started")
	watchedFiles := walk(rootPath)
	for range time.NewTicker(*interval).C {
		scan(rootPath, watchedFiles)
	}
}

func scan(rootPath *string, watchedFiles FileInfo) {
	modifiedDirs := make(map[string]bool, 0)
	for filePath, mostRecentModTime := range walk(rootPath) {
		lastModTime, found := watchedFiles[filePath]
		if found {
			if lastModTime == mostRecentModTime {
				// no changes
				continue
			}

			watchedFiles[filePath] = mostRecentModTime
			if shouldRunTests(filePath, watchedFiles) {
				pkgDir := path.Dir(filePath)
				modifiedDirs[pkgDir] = true
			}
			continue
		}

		// files recently added
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			log.Print("Stat: ", filePath, err)
			continue
		}
		watchedFiles[filePath] = fileInfo.ModTime()
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
		log.Fatal("could not traverse files: ", err)
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

func shouldRunTests(filePath string, watchedFiles FileInfo) bool {
	return isTestFile(filePath) || hasTestFile(filePath, watchedFiles)
}

func isTestFile(fileName string) bool {
	return strings.HasSuffix(fileName, "_test.go")
}

// hasTestFile verifies if a *.go file has a test
func hasTestFile(filePath string, watchedFiles FileInfo) bool {
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
		stdout, _ := exec.Command(
			"go", "test", "-v", "-cover", dir).CombinedOutput()
		result := string(stdout)
		prettyPrint(result)
		notifier.Notify(result, filepath.Base(dir))
	}
}

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func prettyPrint(result string) {
	for _, line := range strings.Split(result, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "--- PASS"):
			pass.Println(trimmed)
		case strings.HasPrefix(trimmed, "--- FAIL"):
			fail.Println(trimmed)
		default:
			fmt.Println(trimmed)
		}
	}
}
