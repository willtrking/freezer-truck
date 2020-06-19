package main

import "os"

// Create a local directory if it does not exist
func MkdirIfNotExists(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, os.ModePerm)
		return !os.IsExist(err), err
	}

	return false, nil
}
