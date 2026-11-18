//go:build darwin || dragonfly || freebsd || illumos || linux || netbsd || openbsd

// Copyright Mycophonic.
// Copyright Forkcloser.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

package xos

import "os"

// ReadDir reads the named directory and returns its entries sorted by name.
// It is os.ReadDir, which opens the path as a directory: opened as a file,
// one that is none is refused for another reason than os gives (a file
// that cannot be read), or waited on (a FIFO with no writer).
//
//nolint:wrapcheck // Thin wrapper
func ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}
