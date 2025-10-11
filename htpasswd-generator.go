// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import (
	"context"

	"github.com/foomo/htpasswd"
	"github.com/golang/glog"
)

type HtpasswdGenerator interface {
	Generate(ctx context.Context, key Key) ([]byte, error)
}

func NewHtpasswdGenerator(connector Connector) HtpasswdGenerator {
	return &htpasswdGenerator{
		connector: connector,
	}
}

type htpasswdGenerator struct {
	connector Connector
}

func (c *htpasswdGenerator) Generate(ctx context.Context, key Key) ([]byte, error) {
	pass, err := c.connector.Password(ctx, key)
	if err != nil {
		glog.V(2).Infof("get password from teamvault for key %v failed: %v", key, err)
		return nil, err
	}
	user, err := c.connector.User(ctx, key)
	if err != nil {
		glog.V(2).Infof("get user from teamvault for key %v failed: %v", key, err)
		return nil, err
	}
	pws := make(htpasswd.HashedPasswords)
	err = pws.SetPassword(string(user), string(pass), htpasswd.HashBCrypt)
	if err != nil {
		glog.V(2).Infof("set password failed for key %v failed: %v", key, err)
		return nil, err
	}
	return pws.Bytes(), nil
}
