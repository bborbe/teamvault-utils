package main

import (
	"os"
	"github.com/bborbe/log"
	io_util "github.com/bborbe/io/util"
	"flag"
	"io"
	"github.com/bborbe/kubernetes_tools/config_parser"
)

const (
	PARAMETER_LOGLEVEL = "loglevel"
	PARAMETER_CONFIG = "config"
)

var (
	logger = log.DefaultLogger
	configPtr = flag.String(PARAMETER_CONFIG, "", "config json file")
	logLevelPtr = flag.String(PARAMETER_LOGLEVEL, log.INFO_STRING, log.FLAG_USAGE)
)

func main() {
	defer logger.Close()
	flag.Parse()

	logger.SetLevelThreshold(log.LogStringToLevel(*logLevelPtr))
	logger.Debugf("set log level to %s", *logLevelPtr)

	err := do(os.Stdout, *configPtr)
	if err != nil {
		logger.Fatal(err)
		logger.Close()
		os.Exit(1)
	}
}

func do(writer io.Writer, configPath string) (error) {
	logger.Debugf("config: %s", configPath)
	configPath, err := io_util.NormalizePath(configPath)
	if err != nil {
		logger.Warnf("normalize path '%s' failed", configPath)
		return err
	}

	configParser := config_parser.New()
	config, err := configParser.ParseConfig(configPath)
	if err != nil {
		logger.Warnf("parse config '%s' failed: %v", config, err)
		return err
	}

	return nil
}
