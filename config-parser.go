// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import (
	"bytes"
	"context"
	"encoding/base64"
	"os"
	"strings"
	"text/template"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

//counterfeiter:generate -o  mocks/config_parser.go --fake-name ConfigParser . ConfigParser
type ConfigParser interface {
	Parse(ctx context.Context, content []byte) ([]byte, error)
}

func NewConfigParser(
	teamvaultConnector Connector,
) ConfigParser {
	return &configParser{
		teamvaultConnector: teamvaultConnector,
	}
}

type configParser struct {
	teamvaultConnector Connector
}

func (c *configParser) Parse(ctx context.Context, content []byte) ([]byte, error) {
	t, err := template.New("config").Funcs(c.createFuncMap(ctx)).Parse(string(content))
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

func (c *configParser) createFuncMap(ctx context.Context) template.FuncMap {
	return template.FuncMap{
		"indent": func(spaces int, v string) string {
			pad := strings.Repeat(" ", spaces)
			return pad + strings.Replace(v, "\n", "\n"+pad, -1)
		},
		"readfile": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("read file for %v", val)
			switch v := val.(type) {
			case string:
				// #nosec G304 -- file path provided by template user
				content, err := os.ReadFile(v)
				if err != nil {
					glog.V(2).Infof("read file %v failed: %v", val, err)
					return "", errors.Wrapf(err, "read file %v failed", val)
				}
				glog.V(4).Infof("return value %s", content)
				return string(content), nil
			default:
				return "", nil
			}
		},
		"teamvaultUser": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("get teamvault value for %v", val)
			if val == nil {
				return "", nil
			}
			key := Key(val.(string))
			if err := key.Validate(ctx); err != nil {
				return nil, errors.Wrapf(err, "key '%s' invalid", key)
			}
			user, err := c.teamvaultConnector.User(ctx, key)
			if err != nil {
				glog.V(2).Infof("get user from teamvault for key %v failed: %v", key, err)
				return "", errors.Wrapf(err, "get user from teamvault for key %v failed", key)
			}
			glog.V(4).Infof("return value %s", user.String())
			return user.String(), nil
		},
		"teamvaultPassword": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("get teamvault value for %v", val)
			if val == nil {
				return "", nil
			}
			key := Key(val.(string))
			if err := key.Validate(ctx); err != nil {
				return nil, errors.Wrapf(err, "key '%s' invalid", key)
			}
			pass, err := c.teamvaultConnector.Password(ctx, key)
			if err != nil {
				glog.V(2).Infof("get password from teamvault for key %v failed: %v", key, err)
				return "", errors.Wrapf(err, "get password from teamvault for key %v failed", key)
			}
			glog.V(4).Infof("return value %s", pass.String())
			return pass.String(), nil
		},
		"teamvaultHtpasswd": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("get teamvault value for %v", val)
			if val == nil {
				return "", nil
			}
			htpasswd := NewHtpasswdGenerator(
				c.teamvaultConnector,
			)
			content, err := htpasswd.Generate(ctx, Key(val.(string)))
			if err != nil {
				return "", errors.Wrapf(err, "generate htpasswd failed")
			}
			glog.V(4).Infof("return value %s", string(content))
			return string(content), nil
		},
		"teamvaultUrl": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("get teamvault value for %v", val)
			if val == nil {
				return "", nil
			}
			key := Key(val.(string))
			if err := key.Validate(ctx); err != nil {
				return nil, errors.Wrapf(err, "key '%s' invalid", key)
			}
			pass, err := c.teamvaultConnector.Url(ctx, key)
			if err != nil {
				glog.V(2).Infof("get url from teamvault for key %v failed: %v", key, err)
				return "", errors.Wrapf(err, "get url from teamvault for key %v failed", key)
			}
			glog.V(4).Infof("return value %s", pass.String())
			return pass.String(), nil
		},
		"teamvaultFile": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("get teamvault value for %v", val)
			if val == nil {
				return "", nil
			}
			key := Key(val.(string))
			if err := key.Validate(ctx); err != nil {
				return nil, errors.Wrapf(err, "key '%s' invalid", key)
			}
			file, err := c.teamvaultConnector.File(ctx, key)
			if err != nil {
				glog.V(2).Infof("get file from teamvault for key %v failed: %v", key, err)
				return "", errors.Wrapf(err, "get file from teamvault for key %v failed", key)
			}
			glog.V(4).Infof("return value %s", file.String())
			content, err := file.Content()
			if err != nil {
				return "", errors.Wrapf(
					err,
					"get content from teamvault file for key %v failed",
					key,
				)
			}
			return string(content), nil
		},
		"teamvaultFileBase64": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("get teamvault value for %v", val)
			if val == nil {
				return "", nil
			}
			key := Key(val.(string))
			if err := key.Validate(ctx); err != nil {
				return nil, errors.Wrapf(err, "key '%s' invalid", key)
			}
			file, err := c.teamvaultConnector.File(ctx, key)
			if err != nil {
				glog.V(2).Infof("get file from teamvault for key %v failed: %v", key, err)
				return "", errors.Wrapf(err, "get file from teamvault for key %v failed", key)
			}
			glog.V(4).Infof("return value %s", file.String())
			content, err := file.Content()
			if err != nil {
				return "", errors.Wrapf(err, "get file from teamvault for key %v failed", key)
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
		"lower": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("lower value %v", val)
			if val == nil {
				return "", nil
			}
			return strings.ToLower(val.(string)), nil
		},
		"upper": func(val interface{}) (interface{}, error) {
			glog.V(4).Infof("upper value %v", val)
			if val == nil {
				return "", nil
			}
			return strings.ToUpper(val.(string)), nil
		},
	}
}
