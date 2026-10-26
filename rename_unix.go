//go:build !windows

// Copyright Mycophonic.
// Copyright Forkcloser.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

package xos

import "os"

// Rename renames (moves) oldpath to newpath, as os.Rename does: a newpath
// that exists and is not a directory is replaced, open or not.
//
//nolint:wrapcheck // Thin wrapper
func Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}
