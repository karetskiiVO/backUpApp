package main

import "os"

func clearDirectory(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	err = os.MkdirAll(path, 0666)
	if err != nil {
		return err
	}

	return nil
}
