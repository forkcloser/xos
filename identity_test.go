package xos_test

// The error's identity, beside its text: a caller tells errors apart with
// os.IsNotExist, os.IsExist, os.IsPermission, errors.Is against the
// syscall errno, and errors.As to a *PathError or *LinkError whose Op and
// Path it reads. A wrap added inside a helper keeps the text and breaks
// every one of those, so each is held to os's answer on every platform,
// over every state and every call.

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/forkcloser/xos"
)

// identity is everything a caller can ask of an error without reading its
// text: the three os predicates, the errno it unwraps to, and the typed
// error it unwraps to with its Op and its Path, the directory masked.
func identity(err error, dir string) string {
	if err == nil {
		return "nil"
	}

	var parts []string

	for name, is := range map[string]bool{
		"IsNotExist":   os.IsNotExist(err),
		"IsExist":      os.IsExist(err),
		"IsPermission": os.IsPermission(err),
		"ErrNotExist":  errors.Is(err, fs.ErrNotExist),
		"ErrExist":     errors.Is(err, fs.ErrExist),
		"ErrPermission": errors.Is(
			err,
			fs.ErrPermission,
		),
		"ErrInvalid": errors.Is(err, fs.ErrInvalid),
		"ErrClosed":  errors.Is(err, fs.ErrClosed),
		"ENOENT":     errors.Is(err, syscall.ENOENT),
		"EEXIST":     errors.Is(err, syscall.EEXIST),
		"EISDIR":     errors.Is(err, syscall.EISDIR),
		"ENOTDIR":    errors.Is(err, syscall.ENOTDIR),
		"EACCES":     errors.Is(err, syscall.EACCES),
		"EPERM":      errors.Is(err, syscall.EPERM),
		"EINVAL":     errors.Is(err, syscall.EINVAL),
		"ENOTEMPTY":  errors.Is(err, syscall.ENOTEMPTY),
		"ENAMETOOLONG": errors.Is(
			err,
			syscall.ENAMETOOLONG,
		),
	} {
		if is {
			parts = append(parts, name)
		}
	}

	var (
		pathErr *os.PathError
		linkErr *os.LinkError
		errno   syscall.Errno
	)

	// The directory masked and the digits with it: a temporary name's
	// random part is the one thing the two sides cannot share.
	mask := func(path string) string {
		return digits.ReplaceAllString(strings.ReplaceAll(path, dir, "DIR"), "N")
	}

	switch {
	case errors.As(err, &pathErr):
		parts = append(parts, "PathError{"+pathErr.Op+" "+mask(pathErr.Path)+"}")
	case errors.As(err, &linkErr):
		parts = append(parts, "LinkError{"+linkErr.Op+" "+mask(linkErr.Old)+" "+mask(linkErr.New)+"}")
	default:
		parts = append(parts, "untyped")
	}

	if errors.As(err, &errno) {
		parts = append(parts, "errno")
	}

	// The map's order is not an identity.
	sortStrings(parts)

	return strings.Join(parts, " ")
}

// writable gives a path back the permission its removal needs: a file
// made read-only by a case, a directory a case left without its search bit.
func writable(path string) {
	if info, err := os.Lstat(path); err == nil && info.IsDir() {
		_ = os.Chmod(path, 0o755)

		return
	}

	_ = os.Chmod(path, 0o644)
}

func sortStrings(parts []string) {
	for i := 1; i < len(parts); i++ {
		for j := i; j > 0 && parts[j] < parts[j-1]; j-- {
			parts[j], parts[j-1] = parts[j-1], parts[j]
		}
	}
}

// calls are every function of the package with its os namesake, each run on
// one path and judged by its error alone.
func calls() map[string][2]func(path string) error {
	closing := func(file *os.File, err error) error {
		if err == nil {
			_ = file.Close()
		}

		return err
	}
	discard := func(_ any, err error) error { return err }

	return map[string][2]func(string) error{
		"Open": {
			func(p string) error { return closing(xos.Open(p)) },
			func(p string) error { return closing(os.Open(p)) },
		},
		"Create": {
			func(p string) error { return closing(xos.Create(p)) },
			func(p string) error { return closing(os.Create(p)) },
		},
		"OpenFile write": {
			func(p string) error { return closing(xos.OpenFile(p, os.O_WRONLY, 0o644)) },
			func(p string) error { return closing(os.OpenFile(p, os.O_WRONLY, 0o644)) },
		},
		"OpenFile exclusive": {
			func(p string) error { return closing(xos.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)) },
			func(p string) error { return closing(os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)) },
		},
		"ReadFile": {
			func(p string) error { return discard(xos.ReadFile(p)) },
			func(p string) error { return discard(os.ReadFile(p)) },
		},
		"ReadDir": {
			func(p string) error { return discard(xos.ReadDir(p)) },
			func(p string) error { return discard(os.ReadDir(p)) },
		},
		"WriteFile": {
			func(p string) error { return xos.WriteFile(p, []byte("w"), 0o644) },
			func(p string) error { return os.WriteFile(p, []byte("w"), 0o644) },
		},
		"Truncate": {
			func(p string) error { return xos.Truncate(p, 1) },
			func(p string) error { return os.Truncate(p, 1) },
		},
		"Stat": {
			func(p string) error { return discard(xos.Stat(p)) },
			func(p string) error { return discard(os.Stat(p)) },
		},
		"Rename onto itself": {
			func(p string) error { return xos.Rename(p, p) },
			func(p string) error { return os.Rename(p, p) },
		},
		"Rename from": {
			func(p string) error { return xos.Rename(p, filepath.Join(filepath.Dir(p), "moved")) },
			func(p string) error { return os.Rename(p, filepath.Join(filepath.Dir(p), "moved")) },
		},
		"Rename onto": {
			func(p string) error { return xos.Rename(filepath.Join(filepath.Dir(p), "absent"), p) },
			func(p string) error { return os.Rename(filepath.Join(filepath.Dir(p), "absent"), p) },
		},
		"CreateTemp in": {
			func(p string) error { return closing(xos.CreateTemp(p, "t*")) },
			func(p string) error { return closing(os.CreateTemp(p, "t*")) },
		},
		"MkdirTemp in": {
			func(p string) error { return discard(xos.MkdirTemp(p, "t*")) },
			func(p string) error { return discard(os.MkdirTemp(p, "t*")) },
		},
		"CreateTemp with a separator in the pattern": {
			func(p string) error { return closing(xos.CreateTemp(filepath.Dir(p), "a/b")) },
			func(p string) error { return closing(os.CreateTemp(filepath.Dir(p), "a/b")) },
		},
		"MkdirTemp with a separator in the pattern": {
			func(p string) error { return discard(xos.MkdirTemp(filepath.Dir(p), "a/b")) },
			func(p string) error { return discard(os.MkdirTemp(filepath.Dir(p), "a/b")) },
		},
	}
}

func TestErrorIdentityIsOs(t *testing.T) {
	t.Parallel()

	for _, start := range states(t) {
		for name, call := range calls() {
			ours, theirs := pair(t, start.setup)

			got := identity(call[0](ours), filepath.Dir(ours))
			want := identity(call[1](theirs), filepath.Dir(theirs))

			if got != want {
				t.Errorf("%s on %s: xos %s, os %s", name, start.name, got, want)
			}

			t.Cleanup(func() { writable(ours); writable(theirs) })
		}
	}

	// Paths that name nothing a filesystem has: empty, a NUL byte, and a
	// name longer than any filesystem takes.
	for name, path := range map[string]string{
		"empty":    "",
		"NUL":      "a\x00b",
		"too long": strings.Repeat("n", 300),
	} {
		for callName, call := range calls() {
			if strings.HasPrefix(callName, "CreateTemp") || strings.HasPrefix(callName, "MkdirTemp") {
				continue
			}

			target := path
			if name != "empty" {
				target = filepath.Join(t.TempDir(), path)
			}

			if got, want := identity(call[0](target), ""), identity(call[1](target), ""); got != want {
				t.Errorf("%s on %s path: xos %s, os %s", callName, name, got, want)
			}
		}
	}
}

// The file a caller gets back is os's file: its Name is the path given,
// and a use after Close answers as os's file does, which on Unix is
// ErrClosed and on Windows the raw handle's refusal; the two are compared,
// not assumed.
func TestFileIsOs(t *testing.T) {
	t.Parallel()

	for name, open := range map[string][2]func(string) (*os.File, error){
		"Open":   {xos.Open, os.Open},
		"Create": {xos.Create, os.Create},
		"OpenFile": {
			func(p string) (*os.File, error) { return xos.OpenFile(p, os.O_RDWR|os.O_CREATE, 0o644) },
			func(p string) (*os.File, error) { return os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0o644) },
		},
	} {
		ours, theirs := pair(t, func(t *testing.T, path string) {
			t.Helper()

			if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		})

		got, err := open[0](ours)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		want, err := open[1](theirs)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		if got.Name() != ours || want.Name() != theirs {
			t.Errorf("%s: Name() = %q / %q, want the paths given", name, got.Name(), want.Name())
		}

		if err := got.Close(); err != nil {
			t.Fatal(err)
		}

		if err := want.Close(); err != nil {
			t.Fatal(err)
		}

		// After Close: a Stat and a second Close, each as os answers it.
		_, gotStat := got.Stat()
		_, wantStat := want.Stat()

		if g, w := identity(gotStat, filepath.Dir(ours)), identity(wantStat, filepath.Dir(theirs)); g != w {
			t.Errorf("%s: Stat after Close: xos %s, os %s", name, g, w)
		}

		if g, w := errorSeen(gotStat, filepath.Dir(ours)), errorSeen(wantStat, filepath.Dir(theirs)); g != w {
			t.Errorf("%s: Stat after Close: xos %q, os %q", name, g, w)
		}

		if g, w := identity(got.Close(), filepath.Dir(ours)), identity(want.Close(), filepath.Dir(theirs)); g != w {
			t.Errorf("%s: a second Close: xos %s, os %s", name, g, w)
		}
	}
}
