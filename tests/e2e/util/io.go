package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(filepath.Clean(src))
	if err != nil {
		return 0, err
	}
	defer func() {
		cerr := source.Close()
		if err == nil {
			err = cerr
		}
	}()

	destination, err := os.Create(filepath.Clean(dst))
	if err != nil {
		return 0, err
	}
	defer func() {
		cerr := destination.Close()
		if err == nil {
			err = cerr
		}
	}()

	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func WriteFile(path string, body []byte) error {
	_, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, body, 0o600)
}
