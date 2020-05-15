// +build !windows

package atomicfile

import (
	"os"
)

// AtomicRename atomically renames (moves) oldpath to newpath.
// It is guaranteed to either replace the target file entirely, or not
// change either file.
func AtomicRename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}
