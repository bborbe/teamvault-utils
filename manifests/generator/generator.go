package generator

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/template"
	"github.com/bborbe/teamvault_utils/manifests/model"
	"github.com/golang/glog"
)

type configGenerator struct {
	userForKey     userForKey
	passwordForKey passwordForKey
	urlForKey      urlForKey
	fileForKey     fileForKey
}

type userForKey func(key model.TeamvaultKey) (model.TeamvaultUser, error)
type passwordForKey func(key model.TeamvaultKey) (model.TeamvaultPassword, error)
type urlForKey func(key model.TeamvaultKey) (model.TeamvaultUrl, error)
type fileForKey func(key model.TeamvaultKey) (model.TeamvaultFile, error)

func New(
	userForKey userForKey,
	passwordForKey passwordForKey,
	urlForKey urlForKey,
	fileForKey fileForKey,
) *configGenerator {
	c := new(configGenerator)
	c.userForKey = userForKey
	c.passwordForKey = passwordForKey
	c.urlForKey = urlForKey
	c.fileForKey = fileForKey
	return c
}

func (c *configGenerator) Generate(sourceDirectory model.SourceDirectory, targetDirectory model.TargetDirectory) error {
	glog.V(4).Infof("generate config from %s to %s", sourceDirectory.String(), targetDirectory.String())
	return filepath.Walk(sourceDirectory.String(),
		func(path string, info os.FileInfo, err error) error {
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
			content, err = c.replaceContent(content)
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

func (c *configGenerator) replaceContent(content []byte) ([]byte, error) {
	funcs := template.FuncMap{
		"teamvaultUser": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("get teamvault value for %v", val)
			if val == nil {
				return "", nil
			}
			pass, err := c.userForKey(model.TeamvaultKey(val.(string)))
			if err != nil {
				glog.V(2).Infof("get user from teamvault failed: %v", err)
				return "", err
			}
			glog.V(4).Infof("return value %s", pass.String())
			return pass.String(), nil
		},
		"teamvaultPassword": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("get teamvault value for %v", val)
			if val == nil {
				return "", nil
			}
			pass, err := c.passwordForKey(model.TeamvaultKey(val.(string)))
			if err != nil {
				glog.V(2).Infof("get password from teamvault failed: %v", err)
				return "", err
			}
			glog.V(4).Infof("return value %s", pass.String())
			return pass.String(), nil
		},
		"teamvaultUrl": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("get teamvault value for %v", val)
			if val == nil {
				return "", nil
			}
			pass, err := c.urlForKey(model.TeamvaultKey(val.(string)))
			if err != nil {
				glog.V(2).Infof("get url from teamvault failed: %v", err)
				return "", err
			}
			glog.V(4).Infof("return value %s", pass.String())
			return pass.String(), nil
		},
		"teamvaultFile": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("get teamvault value for %v", val)
			if val == nil {
				return "", nil
			}
			file, err := c.fileForKey(model.TeamvaultKey(val.(string)))
			if err != nil {
				glog.V(2).Infof("get file from teamvault failed: %v", err)
				return "", err
			}
			glog.V(4).Infof("return value %s", file.String())
			content, err := file.Content()
			if err != nil {
				return "", err
			}
			return string(content), nil
		},
		"teamvaultFileBase64": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("get teamvault value for %v", val)
			if val == nil {
				return "", nil
			}
			file, err := c.fileForKey(model.TeamvaultKey(val.(string)))
			if err != nil {
				glog.V(2).Infof("get file from teamvault failed: %v", err)
				return "", err
			}
			glog.V(4).Infof("return value %s", file.String())
			content, err := file.Content()
			if err != nil {
				return "", err
			}
			return base64.StdEncoding.EncodeToString(content), nil
		},
		"env": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("get env value for %v", val)
			if val == nil {
				return "", nil
			}
			value := os.Getenv(val.(string))
			glog.V(4).Infof("return value %s", value)
			return value, nil
		},
		"base64": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("base64 value %v", val)
			if val == nil {
				return "", nil
			}
			return base64.StdEncoding.EncodeToString([]byte(val.(string))), nil
		},
	}
	t, err := template.New("config").Funcs(funcs).Parse(string(content))
	if err != nil {
		glog.V(2).Infof("parse config failed: %v", err)
		return nil, err
	}
	b := &bytes.Buffer{}
	if err := t.Execute(b, nil); err != nil {
		glog.V(2).Infof("execute template failed: %v", err)
		return nil, err
	}
	return b.Bytes(), nil
}
