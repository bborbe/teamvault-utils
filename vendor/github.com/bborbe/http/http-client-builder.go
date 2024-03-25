// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/golang/glog"
)

type DialFunc func(ctx context.Context, network, address string) (net.Conn, error)

type HttpClientBuilder interface {
	Build() *http.Client
	BuildRoundTripper() http.RoundTripper
	WithProxy() HttpClientBuilder
	WithoutProxy() HttpClientBuilder
	WithRedirects() HttpClientBuilder
	WithoutRedirects() HttpClientBuilder
	WithTimeout(timeout time.Duration) HttpClientBuilder
	WithDialFunc(dialFunc DialFunc) HttpClientBuilder
}

type httpClientBuilder struct {
	proxy         Proxy
	checkRedirect CheckRedirect
	timeout       time.Duration
	dialFunc      DialFunc
}

type Proxy func(req *http.Request) (*url.URL, error)

type CheckRedirect func(req *http.Request, via []*http.Request) error

func NewClientBuilder() HttpClientBuilder {
	b := new(httpClientBuilder)
	b.WithoutProxy()
	b.WithRedirects()
	b.WithTimeout(30 * time.Second)
	return b
}

func (h *httpClientBuilder) WithTimeout(timeout time.Duration) HttpClientBuilder {
	h.timeout = timeout
	return h
}

func (h *httpClientBuilder) WithDialFunc(dialFunc DialFunc) HttpClientBuilder {
	h.dialFunc = dialFunc
	return h
}

func (h *httpClientBuilder) BuildDialFunc() DialFunc {
	if h.dialFunc != nil {
		return h.dialFunc
	}
	return (&net.Dialer{
		Timeout: h.timeout,
	}).DialContext
}

func (h *httpClientBuilder) BuildRoundTripper() http.RoundTripper {
	if glog.V(5) {
		glog.Infof("build http transport")
	}
	return &http.Transport{
		Proxy:           h.proxy,
		DialContext:     h.BuildDialFunc(),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
}

func (h *httpClientBuilder) Build() *http.Client {
	if glog.V(5) {
		glog.Infof("build http client")
	}
	return &http.Client{
		Transport:     h.BuildRoundTripper(),
		CheckRedirect: h.checkRedirect,
	}
}

func (h *httpClientBuilder) WithProxy() HttpClientBuilder {
	h.proxy = http.ProxyFromEnvironment
	return h
}

func (h *httpClientBuilder) WithoutProxy() HttpClientBuilder {
	h.proxy = nil
	return h
}

func (h *httpClientBuilder) WithRedirects() HttpClientBuilder {
	h.checkRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}
	return h
}

func (h *httpClientBuilder) WithoutRedirects() HttpClientBuilder {
	h.checkRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 1 {
			return errors.New("redirects")
		}
		return nil
	}
	return h
}
