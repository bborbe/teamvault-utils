// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/golang/glog"
)

func FileServer(
	root string,
	prefix string,
) http.Handler {
	return &fileServer{
		root:        root,
		prefix:      prefix,
		defaultFile: "index.html",
	}
}

type fileServer struct {
	root        string
	prefix      string
	defaultFile string
}

// ServeHTTP serves index.html if not found
func (f *fileServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

	/* copied from http.ServeHTTP start */
	upath := req.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		req.URL.Path = upath
	}
	name := path.Clean(upath)
	/* copied from http.ServeHTTP end */

	if strings.HasPrefix(name, f.prefix) {
		name = name[len(f.prefix):]
		if name == "" {
			name = f.defaultFile
		}
	}

	file, err := http.Dir(f.root).Open(name)
	if err != nil && os.IsNotExist(err) {
		glog.V(3).Infof("file '%s' not found => serve %s", name, f.defaultFile)
		http.ServeFile(resp, req, path.Join(f.root, f.defaultFile))
		return
	}
	defer file.Close()
	path := path.Join(f.root, name)
	glog.V(3).Infof("serve file '%s'", path)
	http.ServeFile(resp, req, path)
}
