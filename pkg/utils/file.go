package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// WriteFileAtomic writes data to a file atomically using a temp file and rename
func WriteFileAtomic(filename string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)
	tmpfile, err := ioutil.TempFile(dir, ".tmp-")
	if err != nil {
		return err
	}
	tmpname := tmpfile.Name()

	// Clean up temp file on error
	defer func() {
		if tmpfile != nil {
			tmpfile.Close()
			os.Remove(tmpname)
		}
	}()

	if _, err := tmpfile.Write(data); err != nil {
		return err
	}

	if err := tmpfile.Close(); err != nil {
		return err
	}
	tmpfile = nil // Prevent deferred cleanup

	if err := os.Chmod(tmpname, perm); err != nil {
		return err
	}

	return os.Rename(tmpname, filename)
}
