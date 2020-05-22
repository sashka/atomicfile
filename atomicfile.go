// Package atomicfile wraps os.File to allow atomic file updates.
//
// All writes will go to a temporary file.
// Close the file when you are done writing to atomically make the changes visible.
// Abort to discard all your writes.
// This allows a file to be in a consistent state and never expose an in-progress write.
package atomicfile

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// File behaves like os.File, but does an atomic rename operation at Close.
type File struct {
	*os.File
	path      string
	committed bool
}

// New creates a new temporary file that will replace the file at the given
// path when Closed.
func New(path string, mode os.FileMode) (*File, error) {
	dir, file := filepath.Split(path)
	f, err := ioutil.TempFile(dir, "."+file+"-")
	if err != nil {
		return nil, err
	}

	if err := os.Chmod(f.Name(), mode); err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, err
	}
	return &File{File: f, path: path}, nil
}

// Close the file replacing the original file.
func (f *File) Close() error {
	if err := f.File.Close(); err != nil {
		os.Remove(f.File.Name())
		return err
	}
	// In Windows the files should be closed before doing a Rename.
	if err := AtomicRename(f.Name(), f.path); err != nil {
		return err
	}
	f.committed = true
	return nil
}

// Abort closes the file and removes it, discarding all the changes.
// It's safe to call Abort on a file which is already committed.
func (f *File) Abort() error {
	if f.committed {
		return nil
	}
	if err := f.File.Close(); err != nil {
		os.Remove(f.Name())
		return err
	}
	if err := os.Remove(f.Name()); err != nil {
		return err
	}
	return nil
}
