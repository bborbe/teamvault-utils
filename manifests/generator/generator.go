package generator

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/alecthomas/template"
	"github.com/bborbe/kubernetes_tools/manifests/model"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type configGenerator struct {
	passwordForKey passwordForKey
}

type passwordForKey func(key model.TeamvaultKey) (model.TeamvaultPassword, error)

func New(
	passwordForKey passwordForKey,
) *configGenerator {
	c := new(configGenerator)
	c.passwordForKey = passwordForKey
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
				glog.V(2).Infof("replace variables failed: %v", path, err)
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
		"teamvault": func(val interface{}) (interface{}, error) {
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
