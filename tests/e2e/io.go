package e2e

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

//nolint:unused,deadcode // this is called during e2e tests
func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()

	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

//nolint:unused,deadcode // this is called during e2e tests
func writeFile(path string, body []byte) error {
	_, err := os.Create(path)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, body, 0o644) //nolint:gosec //common cosmos issue, but does not work at 600.
}
