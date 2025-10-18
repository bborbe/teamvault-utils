// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

// Config holds the configuration for connecting to a TeamVault instance.
type Config struct {
	Url          Url      `json:"url"`
	User         User     `json:"user"`
	Password     Password `json:"pass"`
	CacheEnabled bool     `json:"cacheEnabled,omitempty"`
}
