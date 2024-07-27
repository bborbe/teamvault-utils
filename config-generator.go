package teamvault

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
)

type SourceDirectory string

func (s SourceDirectory) String() string {
	return string(s)
}

type TargetDirectory string

func (t TargetDirectory) String() string {
	return string(t)
}

//counterfeiter:generate -o  mocks/config_generator.go --fake-name ConfigGenerator . ConfigGenerator
type ConfigGenerator interface {
	Generate(ctx context.Context, sourceDirectory SourceDirectory, targetDirectory TargetDirectory) error
}

type configGenerator struct {
	configParser ConfigParser
}

func NewConfigGenerator(configParser ConfigParser) ConfigGenerator {
	return &configGenerator{
		configParser: configParser,
	}
}

func (c *configGenerator) Generate(ctx context.Context, sourceDirectory SourceDirectory, targetDirectory TargetDirectory) error {
	glog.V(4).Infof("generate config from %s to %s", sourceDirectory.String(), targetDirectory.String())
	return filepath.Walk(sourceDirectory.String(), func(path string, info os.FileInfo, err error) error {
		glog.V(4).Infof("generate path %s info %v", path, info)
		if err != nil {
			return err
		}
		target := fmt.Sprintf("%s%s", targetDirectory.String(), strings.TrimPrefix(path, sourceDirectory.String()))
		glog.V(2).Infof("target: %s", target)
		if info.IsDir() {
			err := os.MkdirAll(target, 0755)
			if err != nil {
				glog.V(2).Infof("create directory %s failed: %v", target, err)
				return err
			}
			glog.V(4).Infof("directory %s created", target)
			return nil
		}
		content, err := ioutil.ReadFile(path)
		if err != nil {
			glog.V(2).Infof("read file %s failed: %v", path, err)
			return err
		}
		content, err = c.configParser.Parse(ctx, content)
		if err != nil {
			glog.V(2).Infof("replace variables failed: %v", err)
			return err
		}
		if err := ioutil.WriteFile(target, content, 0644); err != nil {
			glog.V(2).Infof("create file %s failed: %v", target, err)
			return err
		}
		glog.V(4).Infof("file %s created", target)
		return nil
	})
}
