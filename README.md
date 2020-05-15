# atomicfile

Package `atomicfile` wraps os.File to allow atomic file updates on Linux, macOS, and on Windows.

All writes will go to a temporary file. Call `Close()` explicitly when you are done writing to atomically rename the file
making the changes visible. Call `Abort()` to discard all your writes.

This allows for a file to always be in a consistent state and never represent an in-progress write.

## Installation

Standard `go get`:

```
$ go get github.com/sashka/atomicfile
```

## Usage

```
import "github.com/sashka/atomicfile"

// Prepare to write a file.
f, err := atomicfile.New(path, 0o666)

// It's safe to call f.Abort() on a successfully closed file.
// Otherwise it's correct to discard the file changes.
defer f.Abort()

// Update the file.
if _, err := f.Write(content); err != nil {
    return err
}

// Make changes visible.
f.Close()
```
