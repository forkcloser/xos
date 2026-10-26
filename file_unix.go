//go:build darwin || dragonfly || freebsd || illumos || linux || netbsd || openbsd

// Copyright Mycophonic.
// Copyright Forkcloser.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

package xos

import "os"

// Open opens a file for reading. On Windows, the handle is opened with
// FILE_SHARE_DELETE so that concurrent rename or deletion is permitted.
// On Unix this is the default behavior.
//
//nolint:wrapcheck // Thin wrapper
func Open(path string) (*os.File, error) {
	return os.Open(path) // #nosec G304 -- Caller controls path
}

// OpenFile opens a file with the given flags and permissions. On Windows,
// the handle is opened with FILE_SHARE_DELETE so that concurrent rename
// or deletion is permitted. On Unix this is the default behavior.
//
//nolint:wrapcheck // Thin wrapper
func OpenFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(path, flag, perm) // #nosec G304 -- Caller controls path
}

// Truncate changes the size of the named file.
// On Unix, this delegates directly to os.Truncate (no file handle is opened).
//
//nolint:wrapcheck // Thin wrapper
func Truncate(path string, size int64) error {
	return os.Truncate(path, size)
}
