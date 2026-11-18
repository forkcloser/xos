//go:build windows

// Copyright Mycophonic.
// Copyright Forkcloser.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

package xos

import (
	"os"
	"slices"
	"strings"
	"syscall"
)

// ReadDir reads the named directory and returns its entries sorted by name.
// Equivalent to os.ReadDir but with FILE_SHARE_DELETE on Windows.
func ReadDir(path string) ([]os.DirEntry, error) {
	file, err := Open(path)
	if err != nil {
		return nil, err
	}

	defer func() { _ = file.Close() }() // read-only: a close error carries nothing

	// os.ReadDir opens the path as a directory, so a file is refused by
	// "open": the same error, from fstat here.
	fi, err := file.Stat()
	if err != nil {
		return nil, err //nolint:wrapcheck // Thin wrapper matching os.ReadDir signature.
	}

	if !fi.IsDir() {
		return nil, &os.PathError{Op: "open", Path: path, Err: syscall.ENOTDIR}
	}

	entries, err := file.ReadDir(-1)

	slices.SortFunc(entries, func(a, b os.DirEntry) int {
		return strings.Compare(a.Name(), b.Name())
	})

	return entries, err //nolint:wrapcheck // Thin wrapper matching os.ReadDir signature.
}
