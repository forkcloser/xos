/*
   Copyright Mycophonic.
   Copyright Forkcloser.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package xos

import "os"

const (
	minReadBuf = 512

	// defaultPerm is the standard Unix permission for os.Create (before umask).
	defaultPerm os.FileMode = 0o666

	// tempFilePerm and tempDirPerm are os.CreateTemp's and os.MkdirTemp's.
	tempFilePerm os.FileMode = 0o600
	tempDirPerm  os.FileMode = 0o700
)
