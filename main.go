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

	"github.com/fatih/color"
	"github.com/fishybell/snitch/platform"
)

// FileInfo represents a file and its modification date
type FileInfo map[string]time.Time

var (
	notifier platform.Notifier
	version  = "v1.2.2"
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

	versionFlag := flag.Bool("v", false, "Print the current version and exit")
	rootPath := flag.String("path", defaultPath, "the root path to be watched")
	interval := flag.Duration("interval", 1*time.Second, "the interval (in seconds) for scanning files")
	quietFlag := flag.Bool("q", true, "Only print failing tests (quiet)")
	notifyFlag := flag.Bool("n", false, "Use system notifications")
	fullFlag := flag.Bool("f", false, "Always run entire build")
	flag.Parse()

	if *versionFlag {
		printVersion()
		return
	}

	if *interval < 0 {
		log.Fatal("invalid interval, must be > 0", *interval)
	}

	if err := os.Chdir(*rootPath); err != nil {
		log.Fatal("could not change directory: ", err)
	}

	log.Print("Snitch started for", *rootPath)
	watchedFiles := walk(rootPath)
	for range time.NewTicker(*interval).C {
		scan(rootPath, watchedFiles, *quietFlag, *notifyFlag, *fullFlag)
	}
}

func printVersion() {
	log.Printf("Current build version: %s", version)
}

func scan(rootPath *string, watchedFiles FileInfo, quiet bool, notify bool, full bool) {
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
	if full {
		dedup = append(dedup, "./...")
	} else {
		for dir := range modifiedDirs {
			dedup = append(dedup, dir)
		}
	}
	test(dedup, quiet, notify)
}

func walk(rootPath *string) FileInfo {
	wf := FileInfo{}
	if err := filepath.Walk(*rootPath, visit(wf)); err != nil {
		log.Print("Error: could not traverse files (trying again later): ", err)
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
	b := isTestFile(filePath) || hasTestFile(filePath, watchedFiles)
	return b
}

func isTestFile(fileName string) bool {
	return strings.HasSuffix(fileName, "_test.go")
}

// hasTestFile verifies if a *.go file has a test
func hasTestFile(filePath string, watchedFiles FileInfo) bool {
	dir := filepath.Dir(filePath)
	for k, _ := range watchedFiles {
		if filepath.Dir(k) == dir {
			if isTestFile(k) {
				return true
			}
		}
	}
	return false
}

func test(dirs []string, quiet bool, notify bool) {
	clear()
	for _, dir := range dirs {
		stdout, _ := exec.Command(
			"go", "test", "-v", "-cover", dir).CombinedOutput()
		result := string(stdout)
		prettyPrint(result, quiet)
		if notify {
			notifier.Notify(result, filepath.Base(dir))
		}
	}
}

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
	pass.Println("---------------")
	fmt.Println(" Running tests")
	pass.Println("---------------")
}

func prettyPrint(result string, quiet bool) {
	for _, line := range strings.Split(result, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "=== RUN"):
			if !quiet {
				fmt.Println(trimmed)
			}
		case strings.HasPrefix(trimmed, "--- PASS"):
			if !quiet {
				pass.Println(trimmed)
			}
		case strings.HasPrefix(trimmed, "--- FAIL"):
			fail.Println(trimmed)
		default:
			fmt.Println(trimmed)
		}
	}
}
