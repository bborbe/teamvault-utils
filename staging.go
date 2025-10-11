// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

type Staging bool

func (s Staging) Bool() bool {
	return bool(s)
}
