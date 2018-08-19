package teamvault

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	io_util "github.com/bborbe/io/util"
	"github.com/golang/glog"
)

type VariableName string

func (v VariableName) String() string {
	return string(v)
}

type Key string

func (t Key) String() string {
	return string(t)
}

type SourceDirectory string

func (s SourceDirectory) String() string {
	return string(s)
}

type TargetDirectory string

func (t TargetDirectory) String() string {
	return string(t)
}

type Staging bool

func (s Staging) Bool() bool {
	return bool(s)
}

type Url string

func (t Url) String() string {
	return string(t)
}

type User string

func (t User) String() string {
	return string(t)
}

type Password string

func (t Password) String() string {
	return string(t)
}

type TeamvaultCurrentRevision string

func (t TeamvaultCurrentRevision) String() string {
	return string(t)
}

type File string

func (t File) String() string {
	return string(t)
}

func (t File) Content() ([]byte, error) {
	return base64.StdEncoding.DecodeString(t.String())
}

type TeamvaultConfig struct {
	Url      Url      `json:"url"`
	User     User     `json:"user"`
	Password Password `json:"pass"`
}

type TeamvaultConfigPath string

func (t TeamvaultConfigPath) String() string {
	return string(t)
}

func (d TeamvaultConfigPath) NormalizePath() (TeamvaultConfigPath, error) {
	root, err := io_util.NormalizePath(d.String())
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

func (t TeamvaultConfigPath) Parse() (*TeamvaultConfig, error) {
	path, err := t.NormalizePath()
	if err != nil {
		glog.V(2).Infof("normalize path failed: %v", err)
		return nil, err
	}
	content, err := ioutil.ReadFile(path.String())
	if err != nil {
		glog.Warningf("read config from file %v failed: %v", t, err)
		return nil, err
	}
	return ParseTeamvaultConfig(content)
}

func ParseTeamvaultConfig(content []byte) (*TeamvaultConfig, error) {
	config := &TeamvaultConfig{}
	if err := json.Unmarshal(content, config); err != nil {
		glog.Warningf("parse config failed: %v", err)
		return nil, err
	}
	return config, nil
}

type TeamvaultApiUrl string

func (t TeamvaultApiUrl) String() string {
	return string(t)
}

func (t TeamvaultApiUrl) Key() (Key, error) {
	parts := strings.Split(t.String(), "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("parse key form api-url failed")
	}
	return Key(parts[len(parts)-2]), nil
}
