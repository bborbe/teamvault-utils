// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import "encoding/base64"

// File represents a base64-encoded file stored in TeamVault.
type File string

// String returns the string representation of the File.
func (t File) String() string {
	return string(t)
}

// Content decodes and returns the file content from base64 encoding.
func (t File) Content() ([]byte, error) {
	return base64.StdEncoding.DecodeString(t.String())
}
