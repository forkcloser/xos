// Not illumos: Go's syscall has no Mkfifo there.
//
//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

// Copyright Forkcloser.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

package xos_test

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/forkcloser/xos"
)

// TestReadDirOfWhatIsNoDirectory: ReadDir of a file that cannot be read, of
// a FIFO nothing writes to and of a socket is refused as os.ReadDir refuses
// it, as no directory, and at once. Opened as a file first, the first was
// "permission denied", the second waited for a writer and the third was
// "operation not supported on socket".
func TestReadDirOfWhatIsNoDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	unreadable := filepath.Join(dir, "unreadable")
	if err := os.WriteFile(unreadable, []byte("x"), 0o000); err != nil {
		t.Fatal(err)
	}

	fifo := filepath.Join(dir, "fifo")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}

	// A short name: a socket's address has some hundred bytes for its path.
	socket := filepath.Join(dir, "s")

	var config net.ListenConfig

	listener, err := config.Listen(t.Context(), "unix", socket)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = listener.Close() })

	for _, path := range []string{unreadable, fifo, socket} {
		_, want := os.ReadDir(path)

		got := make(chan error, 1)

		go func() {
			_, err := xos.ReadDir(path)
			got <- err
		}()

		select {
		case err := <-got:
			if !errors.Is(err, syscall.ENOTDIR) || err.Error() != want.Error() {
				t.Errorf("ReadDir(%s) = %v, want %v", filepath.Base(path), err, want)
			}
		case <-time.After(10 * time.Second):
			t.Errorf(
				"ReadDir(%s) has not returned after ten seconds; os.ReadDir returned %v",
				filepath.Base(path),
				want,
			)

			// Let the waiting open go: give the FIFO its other end.
			if writer, err := os.OpenFile(path, os.O_WRONLY|syscall.O_NONBLOCK, 0); err == nil {
				_ = writer.Close()
			}
		}
	}
}
