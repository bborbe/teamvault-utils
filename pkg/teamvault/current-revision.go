// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

// CurrentRevision represents the current revision identifier of a TeamVault secret.
type CurrentRevision string

// String returns the string representation of the CurrentRevision.
func (t CurrentRevision) String() string {
	return string(t)
}
