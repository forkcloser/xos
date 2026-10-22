package xos_test

// The one stated difference, as a whole life of a held file: opened through
// os and through xos, each of the four open calls, then renamed over,
// renamed away, removed, and read by the holder throughout. On every
// platform but Windows the two columns are the same; on Windows the os
// column is what os does there (the rename over and the removal refused
// while the handle is open) and the xos column is the Unix column, which
// is the package's reason to exist.

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/forkcloser/xos"
)

// opens are the four calls that hold a handle.
func opens(xo bool) map[string]func(string) (*os.File, error) {
	if xo {
		return map[string]func(string) (*os.File, error){
			"Open":   xos.Open,
			"Create": func(p string) (*os.File, error) { return xos.OpenFile(p, os.O_RDWR, 0o644) },
			"OpenFile read": func(p string) (*os.File, error) {
				return xos.OpenFile(p, os.O_RDONLY, 0)
			},
			"OpenFile append": func(p string) (*os.File, error) {
				return xos.OpenFile(p, os.O_WRONLY|os.O_APPEND, 0o644)
			},
		}
	}

	return map[string]func(string) (*os.File, error){
		"Open":   os.Open,
		"Create": func(p string) (*os.File, error) { return os.OpenFile(p, os.O_RDWR, 0o644) },
		"OpenFile read": func(p string) (*os.File, error) {
			return os.OpenFile(p, os.O_RDONLY, 0)
		},
		"OpenFile append": func(p string) (*os.File, error) {
			return os.OpenFile(p, os.O_WRONLY|os.O_APPEND, 0o644)
		},
	}
}

// life runs the held file's life and reports each step's outcome; the
// holder reads what it opened only when it opened for reading.
func life(t *testing.T, open func(string) (*os.File, error), rename func(string, string) error, readable bool) string {
	t.Helper()

	dir := t.TempDir()
	held := filepath.Join(dir, "held")
	incoming := filepath.Join(dir, "incoming")

	for path, data := range map[string]string{held: "held", incoming: "incoming"} {
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	holder, err := open(held)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = holder.Close() }()

	var steps []string

	step := func(name string, err error) {
		steps = append(steps, name+"="+errorKind(err))
	}

	// Renamed over: the holder keeps the file it opened.
	step("rename over", rename(incoming, held))

	// What the path holds, its content alone: the mode is the platform's.
	_, content, _ := strings.Cut(onDisk(held), " ")
	steps = append(steps, "path holds "+content)
	// Renamed away, by os: what the holder's handle permits.
	step("rename away", os.Rename(held, filepath.Join(dir, "away")))
	// Removed, by os.
	step("remove", os.Remove(filepath.Join(dir, "away")))
	step("remove original", os.Remove(held))

	if readable {
		if _, seekErr := holder.Seek(0, io.SeekStart); seekErr != nil {
			t.Fatal(seekErr)
		}

		data, readErr := io.ReadAll(holder)
		if readErr != nil {
			t.Fatal(readErr)
		}

		steps = append(steps, "holder reads "+string(data))
	}

	return strings.Join(steps, "; ")
}

func TestHeldFileLife(t *testing.T) {
	t.Parallel()

	for name, open := range opens(true) {
		readable := name != "OpenFile append"
		got := life(t, open, xos.Rename, readable)
		want := life(t, opens(false)[name], os.Rename, readable)

		if runtime.GOOS != "windows" {
			if got != want {
				t.Errorf("%s: xos %q, os %q", name, got, want)
			}

			continue
		}

		// Windows: the xos column is the Unix column; the os column is
		// os's own, held apart here so a change to either is seen.
		unix := map[string]string{
			"Open":            "rename over=none; path holds incoming; rename away=none; remove=none; remove original=not exist; holder reads held",
			"OpenFile read":   "rename over=none; path holds incoming; rename away=none; remove=none; remove original=not exist; holder reads held",
			"Create":          "rename over=none; path holds incoming; rename away=none; remove=none; remove original=not exist; holder reads held",
			"OpenFile append": "rename over=none; path holds incoming; rename away=none; remove=none; remove original=not exist",
		}

		if got != unix[name] {
			t.Errorf("%s through xos on windows: %q, want Unix's %q", name, got, unix[name])
		}

		if strings.HasPrefix(want, "rename over=none") {
			t.Errorf("%s through os on windows replaced a held file: %q", name, want)
		}
	}
}
