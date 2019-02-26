package teamvault

import (
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

func (d *diskFallback) Password(key Key) (Password, error) {
	kind := "password"
	content, err := d.connector.Password(key)
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

func (d *diskFallback) User(key Key) (User, error) {
	kind := "user"
	content, err := d.connector.User(key)
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

func (d *diskFallback) Url(key Key) (Url, error) {
	kind := "url"
	content, err := d.connector.Url(key)
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

func (d *diskFallback) File(key Key) (File, error) {
	kind := "file"
	content, err := d.connector.File(key)
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

func (d *diskFallback) Search(key string) ([]Key, error) {
	return d.connector.Search(key)
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
	return errors.Wrap(ioutil.WriteFile(cachefile(key, kind), content, 0600), "write cache file failed")
}
