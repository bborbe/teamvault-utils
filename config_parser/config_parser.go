package config_parser

import (
	"github.com/bborbe/kubernetes_tools/config"
	"os"
	"github.com/bborbe/log"
	"encoding/json"
)

var logger = log.DefaultLogger

type config_parser struct {

}

type ConfigParser interface {
	ParseConfig(path string) (*config.Config, error)
}

func New() *config_parser {
	return new(config_parser)
}

func (c *config_parser) ParseConfig(path string) (*config.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		logger.Warnf("open filed failed: %v", err)
		return nil, err
	}
	defer file.Close()
	var config config.Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		logger.Warnf("decode json failed: %v", err)
		return nil, err
	}
	return &config, nil
}
