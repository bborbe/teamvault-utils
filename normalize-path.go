// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
)

func NormalizePath(path string) (string, error) {
	glog.V(4).Infof("NormalizePath %s", path)
	if strings.Index(path, "~/") == 0 {
		home := os.Getenv("HOME")
		if len(home) == 0 {
			glog.V(2).Infof("normalize path failed, enviroment variable HOME missing")
			return "", fmt.Errorf("env HOME not found")
		}
		path = fmt.Sprintf("%s/%s", home, path[2:])
		glog.V(4).Infof("replace ~/ with homedir. new path: %s", path)
	}
	result, err := filepath.Abs(path)
	if err != nil {
		glog.Warningf("get absolute path for %v failed: %v", path, err)
		return "", err
	}
	return result, nil
}
