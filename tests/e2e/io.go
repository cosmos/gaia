package e2e

import (
	"fmt"
	"io"
	"os"
)

// copyFile copy file from src to dst
func copyFile(src, dst string) (int64, error) { //nolint:unparam
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

// writeFile write a byte slice into a file path
// create the file if it doesn't exist
// NOTE: this file can be write and read by everyone
func writeFile(path string, body []byte) error {
	return os.WriteFile(path, body, 0o666) //nolint:gosec
}
