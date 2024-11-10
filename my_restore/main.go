package main

import (
	"fmt"
	"os"

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

func restore (backupdir, workdir string) error {



	return nil
}
