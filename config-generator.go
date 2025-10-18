// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
)

// SourceDirectory represents the source directory path for configuration generation.
type SourceDirectory string

// String returns the string representation of the SourceDirectory.
func (s SourceDirectory) String() string {
	return string(s)
}

// TargetDirectory represents the target directory path for configuration generation.
type TargetDirectory string

// String returns the string representation of the TargetDirectory.
func (t TargetDirectory) String() string {
	return string(t)
}

// ConfigGenerator generates configuration files by parsing templates and replacing TeamVault placeholders.
//
//counterfeiter:generate -o  mocks/config_generator.go --fake-name ConfigGenerator . ConfigGenerator
type ConfigGenerator interface {
	Generate(
		ctx context.Context,
		sourceDirectory SourceDirectory,
		targetDirectory TargetDirectory,
	) error
}

type configGenerator struct {
	configParser ConfigParser
}

// NewConfigGenerator creates a new ConfigGenerator with the given ConfigParser.
func NewConfigGenerator(configParser ConfigParser) ConfigGenerator {
	return &configGenerator{
		configParser: configParser,
	}
}

func (c *configGenerator) Generate(
	ctx context.Context,
	sourceDirectory SourceDirectory,
	targetDirectory TargetDirectory,
) error {
	glog.V(4).
		Infof("generate config from %s to %s", sourceDirectory.String(), targetDirectory.String())
	return filepath.Walk(
		sourceDirectory.String(),
		func(path string, info os.FileInfo, err error) error {
			glog.V(4).Infof("generate path %s info %v", path, info)
			if err != nil {
				return err
			}
			target := fmt.Sprintf(
				"%s%s",
				targetDirectory.String(),
				strings.TrimPrefix(path, sourceDirectory.String()),
			)
			glog.V(2).Infof("target: %s", target)
			if info.IsDir() {
				err := os.MkdirAll(target, 0700)
				if err != nil {
					glog.V(2).Infof("create directory %s failed: %v", target, err)
					return err
				}
				glog.V(4).Infof("directory %s created", target)
				return nil
			}
			// #nosec G304 -- path is controlled by filepath.Walk within sourceDirectory
			content, err := os.ReadFile(path)
			if err != nil {
				glog.V(2).Infof("read file %s failed: %v", path, err)
				return err
			}
			content, err = c.configParser.Parse(ctx, content)
			if err != nil {
				glog.V(2).Infof("replace variables failed: %v", err)
				return err
			}
			if err := os.WriteFile(target, content, 0600); err != nil {
				glog.V(2).Infof("create file %s failed: %v", target, err)
				return err
			}
			glog.V(4).Infof("file %s created", target)
			return nil
		},
	)
}
