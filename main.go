package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/fishybell/snitch/platform"
)

// FileInfo represents a file and its modification date
type FileInfo map[string]time.Time

var (
	notifier       platform.Notifier
	version        = "v1.6.3"
	pass           = color.New(color.FgGreen)
	fail           = color.New(color.FgHiRed)
	termReg        = regexp.MustCompile("[0-9]* ([0-9]*)\n")
	passTestReg    = regexp.MustCompile("^PASS$")
	okTestReg      = regexp.MustCompile("^ok.*[a-zA-Z0-9/-_]*\\s*[0-9]*\\.[0-9]*s$")
	skippedTestReg = regexp.MustCompile("^?\\s*[a-zA-Z0-9/-_]*\\s*\\[no test files\\]$")
	wrongPathReg   = regexp.MustCompile("^?\\s*can't load package: package (.*): import (.*): cannot import absolute path")
	wrongSubDirReg = regexp.MustCompile("^?\\s*can't load package: package (.*) is not.*GOROOT \\((.*)\\)")
	outsideDirReg  = regexp.MustCompile("^?\\s*go: directory (.*) is outside main module")
	topDirMemo     = map[string]string{}
)

func init() {
	notifier = platform.NewNotifier()
}

func main() {
	defaultPath, err := os.Getwd()
	if err != nil {
		log.Fatal("could not get current directory: ", err)
	}

	versionFlag := flag.Bool("v", false, "[v]ersion: Print the current version and exit")
	rootPath := flag.String("path", defaultPath, "The root path to be watched")
	interval := flag.Duration("interval", 1*time.Second, "The interval (in seconds) for scanning files")
	quietFlag := flag.Bool("q", true, "[q]uiet: Only print failing tests (use -q=false to be noisy again)")
	notifyFlag := flag.Bool("n", false, "[n]otify: Use system notifications")
	onceFlag := flag.Bool("o", false, "[o]nce: Only fail once, don't run subsequent tests")
	fullFlag := flag.Bool("f", false, "[f]ull: Always run entire build")
	smartFlag := flag.Bool("s", false, "[s]mart: Run entire build when no test files are found")
	flag.Parse()
	remainder := flag.Args()

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
		scan(rootPath, watchedFiles, *quietFlag, *notifyFlag, *onceFlag, *fullFlag, *smartFlag, remainder)
	}
}

func printVersion() {
	log.Printf("Current build version: %s", version)
}

func scan(rootPath *string, watchedFiles FileInfo, quiet bool, notify bool, once bool, full bool, smart bool, remainder []string) {
	modifiedDirs := make(map[string]bool, 0)
	for filePath, mostRecentModTime := range walk(rootPath) {
		lastModTime, found := watchedFiles[filePath]
		if found {
			if lastModTime == mostRecentModTime {
				// no changes
				continue
			}

			watchedFiles[filePath] = mostRecentModTime
			pkgDir := path.Dir(filePath)
			if shouldRunTests(filePath, watchedFiles) {
				modifiedDirs[pkgDir] = true
			} else {
				if smart {
					fmt.Println("running all tests....")
					modifiedDirs[pkgDir] = true
					full = true
				} else {
					sadClear(filePath)
				}
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
	recurse := false
	curDir, _ := os.Getwd()
	if full {
		clear(" Running all tests")
		dedup = append(dedup, "./...")
		topDir := findTopDir(curDir, modifiedDirs)

		if len(topDir) > 0 {
			os.Chdir(topDir)
		}
	} else {
		clear(" Running tests")
		for dir := range modifiedDirs {
			dedup = append(dedup, dir)
			recurse = true
		}
	}
	test(dedup, quiet, notify, once, remainder, recurse)
	os.Chdir(curDir)
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
	for k := range watchedFiles {
		if filepath.Dir(k) == dir {
			if isTestFile(k) {
				return true
			}
		}
	}
	return false
}

func test(dirs []string, quiet bool, notify bool, once bool, remainder []string, recurse bool) {
	for _, dir := range dirs {
		// build up our command
		args := []string{"test", "-v"}
		if once {
			args = append(args, "-failfast")
		}
		args = append(args, remainder...)
		args = append(args, dir)

		// run it and grab the results
		stdout, _ := exec.Command("go", args...).CombinedOutput()
		result := string(stdout)

		// be nice to our eyeballs, and give us the results we're looking for
		if followPath, recursable := prettyPrint(result, quiet); followPath != "" {
			// it turns out we failed to run the tests completely, and should just change directories and re-run
			if recurse {
				curDir, _ := os.Getwd()
				// try directly first
				newDir := followPath
				err := os.Chdir(newDir)
				if err != nil {
					// next go into it as a subdir
					newDir = path.Join(curDir, followPath)
					err = os.Chdir(newDir)
					if err != nil {
						// finally try it in the vendor dir
						newDir = path.Join(curDir, "vendor", followPath)
						err = os.Chdir(newDir)
					}
				}

				if err == nil {
					test([]string{newDir}, quiet, notify, once, remainder, recursable)
				} else {
					fmt.Println(" Failed to recurse into", followPath, "for tests:\n\t", err)
					fail.Println(fillTerminal())
					fmt.Print(result)
				}
			} else {
				fmt.Println(" Not recursing into", followPath, "for tests")
				fail.Println(fillTerminal())
				fmt.Print(result)
			}
		}
		if notify {
			notifier.Notify(result, filepath.Base(dir))
		}
	}
}

func fillTerminal() string {
	return strings.Repeat("-", termWidth())
}

func termWidth() int {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 80 // standard width, back in the day
	}

	subs := termReg.FindStringSubmatch(string(out))
	if len(subs) != 2 {
		return 80 // standard width, back in the day
	}

	width, err := strconv.Atoi(subs[1]) // the 2nd index will be our sub-match, the 1st will be the whole kit and kaboodle
	if err != nil {
		return 80 // standard width, back in the day
	}

	return width
}

func clear(msg string) {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
	pass.Println(fillTerminal())
	fmt.Println(msg)
	pass.Println(fillTerminal())
}

func sadClear(filePath string) {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
	fail.Println(fillTerminal())
	fmt.Println(" No tests found for file:\n", filePath)
	fmt.Println("\n Use `snitch -s` to instead run all tests")
	fail.Println(fillTerminal())
}

func prettyPrint(result string, quiet bool) (followPath string, recursable bool) {
	// see if we failed to test because we're completely outside of the correct path
	found := wrongPathReg.FindSubmatch([]byte(result))
	if len(found) > 1 {
		followPath = string(found[1])

		return
	}

	// see if we failed to test because we're in the wrong directory
	found = wrongSubDirReg.FindSubmatch([]byte(result))
	if len(found) > 1 {
		followPath = string(found[1])

		return
	}

	// see if we're outside of the main module path
	found = outsideDirReg.FindSubmatch([]byte(result))
	if len(found) > 1 {
		followPath = string(found[1])

		return
	}

	lastTrim := 0
	testsPassed := 0
	testsFailed := 0
	for _, line := range strings.Split(result, "\n") {
		// trim all of it, and see how much got lopped
		fullTrim := strings.TrimSpace(line)
		trimmed := len(line) - len(fullTrim)
		partTrim := fullTrim

		// always chunk off the base trimming
		toTrim := lastTrim
		if lastTrim == 0 {
			lastTrim = trimmed
			toTrim = lastTrim
		}

		// if we need more, trim half the remainder, but a leave a little to increase readability
		if trimmed > toTrim {
			toTrim += (trimmed - toTrim) / 2
		}

		// make the cut (but don't cut off the beginnning)
		if len(line) > toTrim {
			partTrim = line[min(toTrim, trimmed):]
		}

		switch {
		case strings.HasPrefix(fullTrim, "=== RUN"):
			if !quiet {
				fmt.Println(partTrim)
			}
		case strings.HasPrefix(fullTrim, "--- PASS"):
			testsPassed++
			if !quiet {
				pass.Println(partTrim)
			}
		case strings.HasPrefix(fullTrim, "--- FAIL"):
			testsFailed++
			fail.Println(partTrim)
		case passTestReg.Match([]byte(fullTrim)):
			testsPassed++
			if !quiet {
				fmt.Println(partTrim)
			}
		case okTestReg.Match([]byte(fullTrim)):
			if !quiet {
				fmt.Println(partTrim)
			}
		case skippedTestReg.Match([]byte(fullTrim)):
			if !quiet {
				fmt.Println(partTrim)
			}
		default:
			fmt.Println(partTrim)
		}
	}

	if testsPassed > 0 {
		pass.Printf("\nPassed %d tests\n", testsPassed)
	}
	if testsFailed > 0 {
		fail.Printf("\nFailed %d tests\n", testsFailed)
	}

	return
}

func min(a, b int) int {
	if a > b {
		return b
	}

	return a
}

// memoized function to find the deepest directory with a go.mod for a given path
func findTopDir(curDir string, modifiedDirs map[string]bool) string {
	var topDir string
	for dir := range modifiedDirs {
		var candidate string
		// have we already found the go.mod for this directory?
		if value, ok := topDirMemo[dir]; ok {
			candidate = value
		} else {

			// well, I guess we've never found it, so find it
			candidate = path.Dir(dir)
			if candidate == curDir {
				// if we're checking the same directory we're already in, no need to look anywhere else
				return ""
			}

			goFile := path.Join(candidate, "go.mod") // obviously broken if your project doesn't have a go.mod file
			if !fileExists(goFile) {

				// keep trying to find it
				candidate = findTopDir(curDir, map[string]bool{
					candidate: true,
				})
			}

			if len(candidate) > 0 {
				// remember our decision for this path
				topDirMemo[dir] = candidate
			}
		}

		// find the deepest directory only (or at least the one with the longest name) as we're only changing dirs once
		if len(candidate) > len(topDir) {
			topDir = candidate
		}
	}

	return topDir
}

// from https://golangcode.com/check-if-a-file-exists/
// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
