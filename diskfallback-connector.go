// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

func NewDiskFallbackConnector(connector Connector) Connector {
	return &diskFallback{
		connector: connector,
	}

}

type diskFallback struct {
	connector Connector
}

func (d *diskFallback) Password(ctx context.Context, key Key) (Password, error) {
	kind := "password"
	content, err := d.connector.Password(ctx, key)
	if err != nil {
		content, err := read(key, kind)
		if err == nil {
			return Password(content), nil
		}
	}
	if write(key, kind, []byte(content)) != nil {
		glog.Warningf("write teamvault diskfallback failed")
	}
	return content, err
}

func (d *diskFallback) User(ctx context.Context, key Key) (User, error) {
	kind := "user"
	content, err := d.connector.User(ctx, key)
	if err != nil {
		content, err := read(key, kind)
		if err == nil {
			return User(content), nil
		}
	}
	if write(key, kind, []byte(content)) != nil {
		glog.Warningf("write teamvault diskfallback failed")
	}
	return content, err
}

func (d *diskFallback) Url(ctx context.Context, key Key) (Url, error) {
	kind := "url"
	content, err := d.connector.Url(ctx, key)
	if err != nil {
		content, err := read(key, kind)
		if err == nil {
			return Url(content), nil
		}
	}
	if write(key, kind, []byte(content)) != nil {
		glog.Warningf("write teamvault diskfallback failed")
	}
	return content, err
}

func (d *diskFallback) File(ctx context.Context, key Key) (File, error) {
	kind := "file"
	content, err := d.connector.File(ctx, key)
	if err != nil {
		content, err := read(key, kind)
		if err == nil {
			return File(content), nil
		}
	}
	if write(key, kind, []byte(content)) != nil {
		glog.Warningf("write teamvault diskfallback failed")
	}
	return content, err
}

func (d *diskFallback) Search(ctx context.Context, key string) ([]Key, error) {
	return d.connector.Search(ctx, key)
}

func cachefile(key Key, kind string) string {
	return filepath.Join(os.Getenv("HOME"), ".teamvault-cache", key.String(), kind)
}

func cachedir(key Key) string {
	return filepath.Join(os.Getenv("HOME"), ".teamvault-cache", key.String())
}

func read(key Key, kind string) ([]byte, error) {
	return ioutil.ReadFile(cachefile(key, kind))
}

func write(key Key, kind string, content []byte) error {
	err := os.MkdirAll(cachedir(key), 0700)
	if err != nil {
		return errors.Wrap(err, "mkdir %s failed")
	}
	return errors.Wrap(
		ioutil.WriteFile(cachefile(key, kind), content, 0600),
		"write cache file failed",
	)
}
