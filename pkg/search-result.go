// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

// SearchResult is one match from a TeamVault secret search. Name/Username/Url
// come directly from the search response's result object (no per-key fetch).
type SearchResult struct {
	Key      Key
	Name     string
	Username string
	Url      Url
}
