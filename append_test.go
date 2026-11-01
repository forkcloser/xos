// Copyright Mycophonic.
// Copyright Forkcloser.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

package xos_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forkcloser/xos"
)

// TestWriteAtOnAppend: WriteAt on a file opened with O_APPEND is refused as
// os refuses it, on every platform. On Windows xos makes its file with
// os.NewFile and sets os's append mark itself; this is what proves the mark
// took, and what fails first when a Go release moves the field.
func TestWriteAtOnAppend(t *testing.T) {
	t.Parallel()

	run := func(open func(string, int, os.FileMode) (*os.File, error)) (string, string) {
		path := filepath.Join(t.TempDir(), "log")
		if err := os.WriteFile(path, []byte("start"), 0o644); err != nil {
			t.Fatal(err)
		}

		file, err := open(path, os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			t.Fatal(err)
		}

		_, writeErr := file.WriteAt([]byte("+"), 0)

		if err := file.Close(); err != nil {
			t.Fatal(err)
		}

		_, content, _ := cutMode(onDisk(path))

		return errorSeen(writeErr, filepath.Dir(path)), content
	}

	gotErr, gotContent := run(xos.OpenFile)
	wantErr, wantContent := run(os.OpenFile)

	if wantErr == "none" || wantContent != "start" {
		t.Fatalf("os accepted WriteAt on an O_APPEND file: %s, %q", wantErr, wantContent)
	}

	if gotErr != wantErr || gotContent != wantContent {
		t.Fatalf("xos %s %q, os %s %q", gotErr, gotContent, wantErr, wantContent)
	}
}

// cutMode splits onDisk's "mode content" into its two parts.
func cutMode(seen string) (mode, content string, ok bool) {
	for i := range len(seen) {
		if seen[i] == ' ' {
			return seen[:i], seen[i+1:], true
		}
	}

	return seen, "", false
}
