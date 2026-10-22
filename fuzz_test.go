package xos_test

// Past the bounded states: arbitrary names under a scratch directory, through
// every call beside its os namesake, the error's identity and text and what
// is on disk compared; and arbitrary temporary-name patterns.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forkcloser/xos"
)

// inside is name joined under dir, or "" when the join escapes dir or names
// dir itself, which the two sides could not keep apart.
func inside(dir, name string) string {
	path := filepath.Join(dir, name)

	rel, err := filepath.Rel(dir, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return ""
	}

	return path
}

func FuzzPaths(f *testing.F) {
	f.Add("file", "x", uint8(0))
	f.Add("a/b/c", "", uint8(1))
	f.Add("name with spaces.txt", "data", uint8(2))
	f.Add("ünïcödé", "d", uint8(3))
	f.Add("trailing/", "", uint8(4))
	f.Add("..", "", uint8(5))
	f.Add("a\x00b", "", uint8(6))
	f.Add(strings.Repeat("l", 260), "", uint8(7))

	f.Fuzz(func(t *testing.T, name, data string, pick uint8) {
		ours, theirs := t.TempDir(), t.TempDir()

		// The same starting state on both sides: a file where the seed
		// says so.
		for _, dir := range []string{ours, theirs} {
			if pick%2 == 1 {
				if err := os.WriteFile(filepath.Join(dir, "file"), []byte("start"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
		}

		ourPath, theirPath := inside(ours, name), inside(theirs, name)
		if ourPath == "" || theirPath == "" {
			return
		}

		names := make([]string, 0, len(calls()))
		for callName := range calls() {
			names = append(names, callName)
		}

		sortStrings(names)

		callName := names[int(pick/2)%len(names)]
		call := calls()[callName]

		gotErr, wantErr := call[0](ourPath), call[1](theirPath)

		if got, want := identity(gotErr, ours), identity(wantErr, theirs); got != want {
			t.Fatalf("%s(%q): xos %s, os %s", callName, name, got, want)
		}

		if got, want := digits.ReplaceAllString(
			errorSeen(gotErr, ours),
			"N",
		), digits.ReplaceAllString(
			errorSeen(wantErr, theirs),
			"N",
		); got != want {
			t.Fatalf("%s(%q): xos %q, os %q", callName, name, got, want)
		}

		if got, want := onDisk(ourPath), onDisk(theirPath); got != want {
			t.Fatalf("%s(%q): xos left %q, os %q", callName, name, got, want)
		}

		_ = data
	})
}

func FuzzTempPatterns(f *testing.F) {
	f.Add("x*", true)
	f.Add("", false)
	f.Add("a*b*c", true)
	f.Add("a/b", false)
	f.Add("*", true)
	f.Add(strings.Repeat("p", 300), false)

	f.Fuzz(func(t *testing.T, pattern string, file bool) {
		ours, theirs := t.TempDir(), t.TempDir()

		var got, want string

		if file {
			gotFile, gotErr := xos.CreateTemp(ours, pattern)
			wantFile, wantErr := os.CreateTemp(theirs, pattern)

			got, want = describeTemp(gotFile, gotErr, ours), describeTemp(wantFile, wantErr, theirs)
		} else {
			gotName, gotErr := xos.MkdirTemp(ours, pattern)
			wantName, wantErr := os.MkdirTemp(theirs, pattern) //nolint:usetesting // the reference

			got, want = describeTempDir(gotName, gotErr, ours), describeTempDir(wantName, wantErr, theirs)
		}

		if got != want {
			t.Fatalf("pattern %q: xos %q, os %q", pattern, got, want)
		}
	})
}

func describeTemp(file *os.File, err error, dir string) string {
	if err != nil {
		return identity(err, dir) + " " + digits.ReplaceAllString(errorSeen(err, dir), "N")
	}

	defer func() { _ = file.Close() }()

	info, statErr := file.Stat()
	if statErr != nil {
		return "stat: " + errorKind(statErr)
	}

	return tempShape(dir, file.Name()) + " " + info.Mode().String()
}

func describeTempDir(name string, err error, dir string) string {
	if err != nil {
		return identity(err, dir) + " " + digits.ReplaceAllString(errorSeen(err, dir), "N")
	}

	info, statErr := os.Stat(name)
	if statErr != nil {
		return "stat: " + errorKind(statErr)
	}

	return tempShape(dir, name) + " " + info.Mode().String()
}
