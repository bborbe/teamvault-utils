package main

import (
	"flag"
	"fmt"
	"os"

	io_util "github.com/bborbe/io/util"
	"github.com/bborbe/kubernetes_tools/config_parser"
	"github.com/bborbe/kubernetes_tools/generator"
	"github.com/bborbe/kubernetes_tools/model"
	"github.com/bborbe/log"
	"runtime"
)

const (
	PARAMETER_LOGLEVEL = "loglevel"
	PARAMETER_CONFIG   = "config"
)

var (
	logger      = log.DefaultLogger
	configPtr   = flag.String(PARAMETER_CONFIG, "", "config json file")
	logLevelPtr = flag.String(PARAMETER_LOGLEVEL, log.INFO_STRING, log.FLAG_USAGE)
)

func main() {
	defer logger.Close()
	flag.Parse()

	logger.SetLevelThreshold(log.LogStringToLevel(*logLevelPtr))
	logger.Debugf("set log level to %s", *logLevelPtr)

	runtime.GOMAXPROCS(runtime.NumCPU())

	err := do(*configPtr)
	if err != nil {
		logger.Fatal(err)
		logger.Close()
		os.Exit(1)
	}
}

func do(configPath string) error {
	if len(configPath) == 0 {
		return fmt.Errorf("parameter %s missing", PARAMETER_CONFIG)
	}

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

	configWriter := generator.New()
	cluster := model.NewCluster(config)
	if err := configWriter.Write(cluster); err != nil {
		logger.Warnf("write configs failed: %v", err)
		return err
	}

	logger.Debugf("generate kubernetes cluster configs completed")

	return nil
}
