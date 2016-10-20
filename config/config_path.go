package config

import (
	"encoding/json"
	"os"

	io_util "github.com/bborbe/io/util"
	"github.com/golang/glog"
)

type ConfigPath string

func (c ConfigPath) String() string {
	return string(c)
}
func (c ConfigPath) NormalizePath() (ConfigPath, error) {
	path, err := io_util.NormalizePath(c.String())
	if err != nil {
		glog.Warningf("normalize path '%s' failed", path)
		return "", err
	}
	return ConfigPath(path), nil
}

func (c ConfigPath) ParseConfig() (*Cluster, error) {
	file, err := os.Open(c.String())
	if err != nil {
		glog.Warningf("open filed failed: %v", err)
		return nil, err
	}
	defer file.Close()
	var config Cluster
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		glog.Warningf("decode json failed: %v", err)
		return nil, err
	}
	return &config, nil
}
