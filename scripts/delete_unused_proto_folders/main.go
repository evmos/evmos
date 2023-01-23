// This script aims to delete all proto files, that do not have
// the option go_package defined in them.
package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	folder := os.Args[1]

	err := filepath.Walk(folder,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !strings.Contains(path, ".proto") {
				return nil
			}

			optionSet, err := checkIfGoPackageSet(path)
			if err != nil {
				return err
			}

			if !optionSet {
				if err = os.Remove(path); err != nil {
					return err
				}
			}

			return nil
		},
	)

	if err != nil {
		log.Fatalf("error while walking the directory: %s", err)
	}
}

// checkIfGoPackageSet takes in the path to a Protobuf file and checks if
// the option "go_package" is set in the file.
func checkIfGoPackageSet(path string) (bool, error) {
	readFile, err := os.Open(path)
	if err != nil {
		return false, err
	}

	var fileLines []string
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		fileLines = append(fileLines, fileScanner.Text())
	}

	if err = readFile.Close(); err != nil {
		return false, err
	}

	for _, line := range fileLines {
		if strings.Contains(line, "option go_package") {
			return true, nil
		}
	}

	return false, nil
}
