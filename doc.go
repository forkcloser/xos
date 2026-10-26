// Copyright Mycophonic.
// Copyright Forkcloser.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

// Package xos provides the same filesystem features as golang os package, with a key difference on windows
// (FILE_SHARE_DELETE). The one thing os does there that xos does not is the \\?\ prefix os adds to a long
// path on Windows 10 before 1703 and Server 2016; on anything later the Go runtime makes the process
// long-path aware and neither needs it.
package xos
