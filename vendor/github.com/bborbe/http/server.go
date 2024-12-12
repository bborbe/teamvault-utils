// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/bborbe/errors"
	"github.com/bborbe/run"
	"github.com/golang/glog"
)

func NewServerWithPort(port int, router http.Handler) run.Func {
	return NewServer(
		fmt.Sprintf(":%d", port),
		router,
	)
}

func NewServer(addr string, router http.Handler) run.Func {
	return func(ctx context.Context) error {

		server := &http.Server{
			Addr:      addr,
			Handler:   router,
			TLSConfig: nil,
		}
		go func() {
			select {
			case <-ctx.Done():
				if err := server.Shutdown(ctx); err != nil {
					glog.Warningf("shutdown failed: %v", err)
				}
			}
		}()
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			glog.V(0).Info(err)
			return nil
		}
		return errors.Wrapf(ctx, err, "httpServer failed")
	}
}

func NewServerTLS(addr string, router http.Handler, serverCertPath string, serverKeyPath string) run.Func {
	return func(ctx context.Context) error {
		server := &http.Server{
			Addr:     addr,
			Handler:  router,
			ErrorLog: log.New(NewSkipErrorWriter(log.Writer()), "", log.LstdFlags),
		}
		go func() {
			select {
			case <-ctx.Done():
				if err := server.Shutdown(ctx); err != nil {
					glog.Warningf("shutdown failed: %v", err)
				}
			}
		}()
		err := server.ListenAndServeTLS(serverCertPath, serverKeyPath)
		if errors.Is(err, http.ErrServerClosed) {
			glog.V(0).Info(err)
			return nil
		}
		return errors.Wrapf(ctx, err, "httpServer failed")
	}
}

func NewSkipErrorWriter(writer io.Writer) io.Writer {
	return &skipErrorWriter{
		writer: writer,
	}
}

type skipErrorWriter struct {
	writer io.Writer
}

func (s *skipErrorWriter) Write(p []byte) (n int, err error) {
	if bytes.Contains(p, []byte("http: TLS handshake error from")) {
		// skip
		return len(p), nil
	}
	return s.writer.Write(p)
}
