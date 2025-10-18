// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import (
	"encoding/json"
	"os"

	"github.com/golang/glog"
)

// TeamvaultConfigPath represents a path to a TeamVault configuration file.
type TeamvaultConfigPath string

// String returns the string representation of the TeamvaultConfigPath.
func (t TeamvaultConfigPath) String() string {
	return string(t)
}

// NormalizePath converts the TeamvaultConfigPath to an absolute path.
func (t TeamvaultConfigPath) NormalizePath() (TeamvaultConfigPath, error) {
	root, err := NormalizePath(t.String())
	if err != nil {
		return "", err
	}
	return TeamvaultConfigPath(root), nil
}

// Exists checks if the TeamvaultConfigPath points to an existing non-empty file.
func (t TeamvaultConfigPath) Exists() bool {
	path, err := t.NormalizePath()
	if err != nil {
		glog.V(2).Infof("normalize path failed: %v", err)
		return false
	}
	fileInfo, err := os.Stat(path.String())
	if err != nil {
		glog.V(2).Infof("file %v exists => false", t)
		return false
	}
	if fileInfo.Size() == 0 {
		glog.V(2).Infof("file %v empty => false", t)
		return false
	}
	if fileInfo.IsDir() {
		glog.V(2).Infof("file %v is dir => false", t)
		return false
	}
	glog.V(2).Infof("file %v exists and not empty => true", t)
	return true
}

// Parse reads and parses the TeamVault configuration from the file.
func (t TeamvaultConfigPath) Parse() (*Config, error) {
	path, err := t.NormalizePath()
	if err != nil {
		glog.V(2).Infof("normalize path failed: %v", err)
		return nil, err
	}
	content, err := os.ReadFile(path.String())
	if err != nil {
		glog.Warningf("read config from file %v failed: %v", t, err)
		return nil, err
	}
	return ParseTeamvaultConfig(content)
}

// ParseTeamvaultConfig parses a TeamVault configuration from JSON content.
func ParseTeamvaultConfig(content []byte) (*Config, error) {
	config := &Config{}
	if err := json.Unmarshal(content, config); err != nil {
		glog.Warningf("parse config failed: %v", err)
		return nil, err
	}
	return config, nil
}
