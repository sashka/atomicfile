package atomicfile

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"runtime"
	"testing"
)

func test(t *testing.T, dir, prefix string) {
	t.Parallel()

	tmpfile, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		t.Fatal(err)
	}

	name := tmpfile.Name()

	// In Windows the files should be closed before doing a Remove.
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(name); err != nil {
		t.Fatal(err)
	}

	defer os.Remove(name)

	f, err := New(name, os.FileMode(0o666))
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("foo"))
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		t.Fatal("did not expect file to exist")
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(name); err != nil {
		t.Fatalf("expected file to exist: %s", err)
	}
}

func TestCurrentDir(t *testing.T) {
	cwd, _ := os.Getwd()
	test(t, cwd, "atomicfile-current-dir-")
}

func TestRootTmpDir(t *testing.T) {
	if runtime.GOOS != "windows" {
		test(t, "/tmp", "atomicfile-root-tmp-dir-")
	}
}

func TestDefaultTmpDir(t *testing.T) {
	test(t, "", "atomicfile-default-tmp-dir-")
}

func TestAbort(t *testing.T) {
	contents := []byte("the answer is 42")
	t.Parallel()
	tmpfile, err := ioutil.TempFile("", "atomicfile-abort-")
	if err != nil {
		t.Fatal(err)
	}
	name := tmpfile.Name()
	if _, err := tmpfile.Write(contents); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(name)

	f, err := New(name, os.FileMode(0o666))
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("foo"))
	if err := f.Abort(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(name); err != nil {
		t.Fatalf("expected file to exist: %s", err)
	}
	actual, err := ioutil.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(contents, actual) {
		t.Fatalf(`did not find expected "%s" instead found "%s"`, contents, actual)
	}
}

func TestDoubleClose(t *testing.T) {
	contents := []byte("the answer is 42")
	t.Parallel()
	tmpfile, err := ioutil.TempFile("", "atomicfile-double-")
	if err != nil {
		t.Fatal(err)
	}
	name := tmpfile.Name()
	if _, err := tmpfile.Write(contents); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(name)

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); !errors.Is(err, os.ErrClosed) {
		t.Fatal(err)
	}
}

func TestTypicalHappyPath(t *testing.T) {
	contents := []byte("the answer is 42")
	foo := []byte("foo")

	t.Parallel()
	tmpfile, err := ioutil.TempFile("", "atomicfile-happy-")
	if err != nil {
		t.Fatal(err)
	}
	name := tmpfile.Name()
	if _, err := tmpfile.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(name)

	// Typical happy path
	func(name string) {
		f, err := New(name, os.FileMode(0o666))
		if err != nil {
			t.Fatal(err)
		}
		// f.Abort shouldn't return any error due to changes already committed with f.Close()
		defer func() {
			if err := f.Abort(); err != nil {
				t.Fatal(err)
			}
		}()

		if _, err := f.Write(foo); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}(name)

	if _, err := os.Stat(name); err != nil {
		t.Fatalf("expected file to exist: %s", err)
	}
	actual, err := ioutil.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(foo, actual) {
		t.Fatalf(`did not find expected "%s" instead found "%s"`, contents, actual)
	}
}

func TestTypicalUnhappy(t *testing.T) {
	contents := []byte("the answer is 42")
	foo := []byte("foo")

	t.Parallel()
	tmpfile, err := ioutil.TempFile("", "atomicfile-unhappy-")
	if err != nil {
		t.Fatal(err)
	}
	name := tmpfile.Name()
	if _, err := tmpfile.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(name)

	// Typical unhappy path
	func(name string) {
		f, err := New(name, os.FileMode(0o666))
		if err != nil {
			t.Fatal(err)
		}
		// f.Abort shouldn't successfully remove inner temporary file.
		defer func() {
			if err := f.Abort(); err != nil {
				t.Fatal(err)
			}
		}()

		if _, err := f.Write(foo); err != nil {
			t.Fatal(err)
		}
		// File haven't properly closed for any reason.
	}(name)

	if _, err := os.Stat(name); err != nil {
		t.Fatalf("expected file to exist: %s", err)
	}
	actual, err := ioutil.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(contents, actual) {
		t.Fatalf(`did not find expected "%s" instead found "%s"`, contents, actual)
	}
}
