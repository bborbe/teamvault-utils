// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import "encoding/base64"

type File string

func (t File) String() string {
	return string(t)
}

func (t File) Content() ([]byte, error) {
	return base64.StdEncoding.DecodeString(t.String())
}
