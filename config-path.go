package teamvault

import (
	"encoding/json"
	"os"

	io_util "github.com/bborbe/io/util"
	"github.com/golang/glog"
)

type TeamvaultConfigPath string

func (t TeamvaultConfigPath) String() string {
	return string(t)
}

func (t TeamvaultConfigPath) NormalizePath() (TeamvaultConfigPath, error) {
	root, err := io_util.NormalizePath(t.String())
	if err != nil {
		return "", err
	}
	return TeamvaultConfigPath(root), nil
}

// Exists the backup
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

func ParseTeamvaultConfig(content []byte) (*Config, error) {
	config := &Config{}
	if err := json.Unmarshal(content, config); err != nil {
		glog.Warningf("parse config failed: %v", err)
		return nil, err
	}
	return config, nil
}
