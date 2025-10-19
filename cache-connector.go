// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import (
	"context"
	"sync"
)

// NewCacheConnector creates a new Connector that caches responses from the underlying connector.
func NewCacheConnector(connector Connector) Connector {
	return &cacheConnector{
		connector: connector,
		passwords: make(map[Key]Password),
		users:     make(map[Key]User),
		urls:      make(map[Key]Url),
		files:     make(map[Key]File),
	}
}

type cacheConnector struct {
	connector Connector
	mu        sync.RWMutex
	passwords map[Key]Password
	users     map[Key]User
	urls      map[Key]Url
	files     map[Key]File
}

func (c *cacheConnector) Password(ctx context.Context, key Key) (Password, error) {
	c.mu.RLock()
	value, ok := c.passwords[key]
	c.mu.RUnlock()
	if ok {
		return value, nil
	}
	value, err := c.connector.Password(ctx, key)
	if err == nil {
		c.mu.Lock()
		c.passwords[key] = value
		c.mu.Unlock()
	}
	return value, err
}

func (c *cacheConnector) User(ctx context.Context, key Key) (User, error) {
	c.mu.RLock()
	value, ok := c.users[key]
	c.mu.RUnlock()
	if ok {
		return value, nil
	}
	value, err := c.connector.User(ctx, key)
	if err == nil {
		c.mu.Lock()
		c.users[key] = value
		c.mu.Unlock()
	}
	return value, err
}

func (c *cacheConnector) Url(ctx context.Context, key Key) (Url, error) {
	c.mu.RLock()
	value, ok := c.urls[key]
	c.mu.RUnlock()
	if ok {
		return value, nil
	}
	value, err := c.connector.Url(ctx, key)
	if err == nil {
		c.mu.Lock()
		c.urls[key] = value
		c.mu.Unlock()
	}
	return value, err
}

func (c *cacheConnector) File(ctx context.Context, key Key) (File, error) {
	c.mu.RLock()
	value, ok := c.files[key]
	c.mu.RUnlock()
	if ok {
		return value, nil
	}
	value, err := c.connector.File(ctx, key)
	if err == nil {
		c.mu.Lock()
		c.files[key] = value
		c.mu.Unlock()
	}
	return value, err
}

func (c *cacheConnector) Search(ctx context.Context, key string) ([]Key, error) {
	return c.connector.Search(ctx, key)
}
