package connector

import (
	"github.com/bborbe/teamvault-utils"
	"io/ioutil"
	"path/filepath"
	"os"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type DiskFallback struct {
	Connector teamvault.Connector
}

func (d *DiskFallback) Password(key teamvault.Key) (teamvault.Password, error) {
	kind := "password"
	content, err := d.Connector.Password(key)
	if err != nil {
		content, err := read(key, kind)
		if err == nil {
			return teamvault.Password(content), nil
		}
	}
	if write(key, kind, []byte(content)) != nil {
		glog.Warningf("write teamvault diskfallback failed")
	}
	return content, err
}

func (d *DiskFallback) User(key teamvault.Key) (teamvault.User, error) {
	kind := "user"
	content, err := d.Connector.User(key)
	if err != nil {
		content, err := read(key, kind)
		if err == nil {
			return teamvault.User(content), nil
		}
	}
	if write(key, kind, []byte(content)) != nil {
		glog.Warningf("write teamvault diskfallback failed")
	}
	return content, err
}

func (d *DiskFallback) Url(key teamvault.Key) (teamvault.Url, error) {
	kind := "url"
	content, err := d.Connector.Url(key)
	if err != nil {
		content, err := read(key, kind)
		if err == nil {
			return teamvault.Url(content), nil
		}
	}
	if write(key, kind, []byte(content)) != nil {
		glog.Warningf("write teamvault diskfallback failed")
	}
	return content, err
}

func (d *DiskFallback) File(key teamvault.Key) (teamvault.File, error) {
	kind := "file"
	content, err := d.Connector.File(key)
	if err != nil {
		content, err := read(key, kind)
		if err == nil {
			return teamvault.File(content), nil
		}
	}
	if write(key, kind, []byte(content)) != nil {
		glog.Warningf("write teamvault diskfallback failed")
	}
	return content, err
}

func (d *DiskFallback) Search(key string) ([]teamvault.Key, error) {
	return d.Connector.Search(key)
}

func cachefile(key teamvault.Key, kind string) (string) {
	return filepath.Join(os.Getenv("HOME"), ".teamvault-cache", key.String(), kind)
}

func cachedir(key teamvault.Key) (string) {
	return filepath.Join(os.Getenv("HOME"), ".teamvault-cache", key.String())
}

func read(key teamvault.Key, kind string) ([]byte, error) {
	return ioutil.ReadFile(cachefile(key, kind))
}

func write(key teamvault.Key, kind string, content []byte) (error) {
	err := os.MkdirAll(cachedir(key), 0700)
	if err != nil {
		return errors.Wrap(err, "mkdir %s failed")
	}
	return errors.Wrap(ioutil.WriteFile(cachefile(key, kind), content, 0600), "write cache file failed")
}
