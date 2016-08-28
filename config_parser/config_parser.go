package config_parser

import (
	"encoding/json"
	"os"

	"github.com/bborbe/kubernetes_tools/config"
	"github.com/golang/glog"
)

type config_parser struct {
}

type ConfigParser interface {
	ParseConfig(path string) (*config.Cluster, error)
}

func New() *config_parser {
	return new(config_parser)
}

func (c *config_parser) ParseConfig(path string) (*config.Cluster, error) {
	file, err := os.Open(path)
	if err != nil {
		glog.Warningf("open filed failed: %v", err)
		return nil, err
	}
	defer file.Close()
	var config config.Cluster
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		glog.Warningf("decode json failed: %v", err)
		return nil, err
	}
	return &config, nil
}
