package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
)

func main() {
	args := os.Args[1:]

	var options struct {
		Args struct {
			BackDir string
			WorkDir string
			Unused  []string
		} `positional-args:"yes" required:"yes"`
	}

	parser := flags.NewParser(&options, flags.Default&(^flags.PrintErrors))

	_, err := parser.ParseArgs(args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if len(options.Args.Unused) != 0 {
		fmt.Printf("undefined sequence of arguments: %v\n", options.Args.Unused)
		os.Exit(1)
	}

	err = restore(options.Args.BackDir, options.Args.WorkDir)

}

func restoreFull(backupDir, workDir string, exceptPath map[string]struct{}) error {
	return filepath.Walk(backupDir, func(path string, info fs.FileInfo, err error) error {
		if path == backupDir || path[len(backupDir)+1:] == ".backupcache" {
			return nil
		}

		bufPath := path[len(backupDir):]

		if _, ok := exceptPath[bufPath]; ok {
			return nil
		}

		workPath := workDir + bufPath

		if info.IsDir() {
			err := os.Mkdir(workPath, info.Mode())
			if err != nil {
				return err
			}
		} else {
			srcfile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcfile.Close()

			dstfile, err := os.OpenFile(workPath, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				return err
			}
			defer dstfile.Close()

			io.Copy(dstfile, srcfile)
			dstfile.Chmod(info.Mode())
		}

		return nil
	})
}

func restore(backupDir, workDir string) error {
	err := clearDirectory(workDir)
	if err != nil {
		return err
	}

	buf, err := os.ReadFile(backupDir + "/.backupcache")
	if err != nil {
		return err
	}
	cachedStrings := strings.Split(string(buf), "\n")

	idx := strings.LastIndexAny(backupDir, "/\\")
	lastFullDir := backupDir[:idx] + cachedStrings[0]

	if lastFullDir != backupDir {
		exceptSet := make(map[string]struct{})
		for _, path := range cachedStrings[1:] {
			exceptSet[path] = struct{}{}
		}

		restoreFull(lastFullDir, workDir, exceptSet)
	}

	restoreFull(backupDir, workDir, map[string]struct{}{})

	return nil
}
