package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
)

func main() {
	args := os.Args[1:]

	var options struct {
		Args struct {
			Mode    string
			WorkDir string
			BackDir string
			Unused  []string
		} `positional-args:"yes" required:"3"`
	}

	parser := flags.NewParser(&options, flags.Default&(^flags.PrintErrors))
	parser.Usage = "[\"full\" | \"incremental\"] "

	_, err := parser.ParseArgs(args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if len(options.Args.Unused) != 0 {
		fmt.Printf("undefined sequence of arguments: %v\n", options.Args.Unused)
		os.Exit(1)
	}

	switch options.Args.Mode {
	case "full":
		err = fullBackup(options.Args.WorkDir, options.Args.BackDir)
	case "incremental":
		err = incrementalBackup(options.Args.WorkDir, options.Args.BackDir)
	default:
		err = fmt.Errorf("undefinedmode: %v", options.Args.Mode)
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func fullBackup(workDir, backupDir string) error {
	currentName := marshalTime(time.Now())
	newDir := backupDir + currentName
	ok, err := exists(newDir)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("the directory %v already exists", newDir)
	}
	err = os.MkdirAll(newDir, 0750)
	if err != nil {
		return err
	}

	cachefile, err := os.OpenFile(newDir+"/.backupcache", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	fmt.Fprint(cachefile, currentName)
	cachefile.Close()

	err = filepath.Walk(workDir, func(path string, info fs.FileInfo, err error) error {
		if path == workDir {
			return nil
		}

		backupPath := newDir + path[len(workDir):]

		if info.IsDir() {
			err := os.Mkdir(backupPath, info.Mode())
			if err != nil {
				return err
			}
		} else {
			srcfile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcfile.Close()
			dstfile, err := os.OpenFile(backupPath, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				return err
			}
			defer dstfile.Close()

			io.Copy(dstfile, srcfile)

			dstfile.Chmod(info.Mode())
		}

		return nil
	})
	if err != nil {
		return err
	}

	cachefile, err = os.OpenFile(backupDir+"/.backupcache", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	fmt.Fprint(cachefile, currentName)
	cachefile.Close()

	return nil
}

func incrementalBackup(workDir, backupDir string) error {
	currentName := marshalTime(time.Now())

	prevBackup, err := os.ReadFile(backupDir + "/.backupcache")
	if err != nil {
		return fmt.Errorf(
			"os returned:%w\ndata about the last full backup is not available, check the existence of the full backup and its access rights", err,
		)
	}

	newDir := backupDir + currentName
	ok, err := exists(newDir)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("the directory %v already exists", newDir)
	}
	err = os.MkdirAll(newDir, 0750)
	if err != nil {
		return err
	}

	prevDir := backupDir + string(prevBackup)
	prevTime, err := unmarshalTime(string(prevBackup)[1:])
	if err != nil {
		return err
	}
	cachefile, err := os.OpenFile(newDir+"/.backupcache", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer cachefile.Close()
	prevCachefile, err := os.Open(prevDir + "/.backupcache")
	if err != nil {
		return err
	}
	defer prevCachefile.Close()
	io.Copy(cachefile, prevCachefile)

	backupDiff := make(map[string]struct {
		Size int64
		Used bool
	})
	err = filepath.Walk(prevDir, func(path string, info fs.FileInfo, err error) error {
		if path == prevDir || path[len(prevDir)+1:] == ".backupcache" {
			return nil
		}

		backupDiff[path[len(prevDir):]] = struct {
			Size int64
			Used bool
		}{
			info.Size(), false,
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = filepath.Walk(workDir, func(path string, info fs.FileInfo, err error) error {
		if path == workDir {
			return nil
		}

		bufPath := path[len(workDir):]
		backupPath := newDir + bufPath

		if prevSize, ok := backupDiff[bufPath]; ok {
			if prevSize.Size == info.Size() && !info.ModTime().After(prevTime) {
				prevSize.Used = true
				backupDiff[bufPath] = prevSize
				return nil
			}
		}

		if info.IsDir() {
			err := os.Mkdir(backupPath, info.Mode())
			if err != nil {
				return err
			}
		} else {
			srcfile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcfile.Close()

			lastIdx := strings.LastIndexAny(backupPath, "/\\")
			err = os.MkdirAll(backupPath[:lastIdx], 0666)
			if err != nil {
				return err
			}

			dstfile, err := os.OpenFile(backupPath, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				return err
			}
			defer dstfile.Close()

			io.Copy(dstfile, srcfile)
			dstfile.Chmod(info.Mode())
		}

		return nil
	})
	if err != nil {
		return err
	}

	for path, val := range backupDiff {
		if val.Used {
			continue
		}

		fmt.Fprintf(cachefile, "\n%v", path)
	}

	return nil
}
