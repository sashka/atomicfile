// +build windows

package atomicfile

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	op = "MoveFileExW"

	// See https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-movefileexw
	moveFileReplaceExisting = 0x1
	moveFileWriteThrough    = 0x8
)

var (
	modkernel32     = syscall.NewLazyDLL("kernel32.dll")
	procMoveFileExW = modkernel32.NewProc(op)
)

func moveFileEx(lpExistingFileName *uint16, lpNewFileName *uint16, dwFlags uint32) (err error) {
	r1, _, e1 := syscall.Syscall(procMoveFileExW.Addr(), 3, uintptr(unsafe.Pointer(lpExistingFileName)), uintptr(unsafe.Pointer(lpNewFileName)), uintptr(dwFlags))
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// AtomicRename atomically renames (moves) oldpath to newpath.
// It is guaranteed to either replace the target file entirely, or not
// change either file.
func AtomicRename(oldpath, newpath string) error {
	src, err := syscall.UTF16PtrFromString(oldpath)
	if err != nil {
		return &os.LinkError{op, oldpath, newpath, err}
	}
	dest, err := syscall.UTF16PtrFromString(newpath)
	if err != nil {
		return &os.LinkError{op, oldpath, newpath, err}
	}
	if err := moveFileEx(src, dest, moveFileReplaceExisting|moveFileWriteThrough); err != nil {
		return &os.LinkError{op, oldpath, newpath, err}
	}
	return nil
}
