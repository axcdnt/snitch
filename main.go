package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// FileInfo ...
type FileInfo struct {
	Path         string
	LastModified time.Time
}

func main() {
	wd, _ := os.Getwd()
	rootPath := flag.String("path", wd, "the root path to be watched")
	flag.Parse()

	watchedFiles := walk(rootPath)
	for _, wf := range watchedFiles {
		fmt.Printf("File info: %v", wf)
	}
}

func walk(rootPath *string) []FileInfo {
	var watchedFiles []FileInfo

	err := filepath.Walk(*rootPath, visit(&watchedFiles))
	if err != nil {
		panic(err)
	}

	return watchedFiles
}

func visit(watchedFiles *[]FileInfo) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		if filepath.Ext(path) == ".go" {
			fileInfo := FileInfo{
				Path:         path,
				LastModified: info.ModTime(),
			}

			*watchedFiles = append(*watchedFiles, fileInfo)
		}

		return nil
	}
}
