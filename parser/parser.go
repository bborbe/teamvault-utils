package parser

import (
	"bytes"
	"encoding/base64"
	"os"
	"github.com/bborbe/teamvault_utils/connector"
	"github.com/bborbe/teamvault_utils/model"
	"github.com/golang/glog"
	"text/template"
)

type Parser interface {
	Parse(content []byte) ([]byte, error)
}

type configParser struct {
	teamvaultConnector connector.Connector
}

func New(
	teamvaultConnector connector.Connector,
) *configParser {
	c := new(configParser)
	c.teamvaultConnector = teamvaultConnector
	return c
}

func (c *configParser) Parse(content []byte) ([]byte, error) {
	t, err := template.New("config").Funcs(c.createFuncMap()).Parse(string(content))
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

func (c *configParser) createFuncMap() template.FuncMap {
	return template.FuncMap{
		"teamvaultUser": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("get teamvault value for %v", val)
			if val == nil {
				return "", nil
			}
			pass, err := c.teamvaultConnector.User(model.TeamvaultKey(val.(string)))
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
			pass, err := c.teamvaultConnector.Password(model.TeamvaultKey(val.(string)))
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
			pass, err := c.teamvaultConnector.Url(model.TeamvaultKey(val.(string)))
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
			file, err := c.teamvaultConnector.File(model.TeamvaultKey(val.(string)))
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
			file, err := c.teamvaultConnector.File(model.TeamvaultKey(val.(string)))
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
}
