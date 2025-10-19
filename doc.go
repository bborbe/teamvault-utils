// Copyright (c) 2018-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package teamvault provides utilities for accessing and managing TeamVault secrets.
//
// TeamVault is a secret management system, and this package offers Go clients for
// retrieving passwords, users, URLs, and files from TeamVault instances. It includes
// various connector implementations for different use cases including remote access,
// caching, disk fallback, and testing.
//
// The package also provides configuration parsing and generation capabilities to
// replace TeamVault placeholders in configuration templates with actual secret values.
package teamvault
