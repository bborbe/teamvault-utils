// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"crypto/tls"
	stderrors "errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/bborbe/errors"
	"github.com/golang/glog"
)

type DialFunc func(ctx context.Context, network, address string) (net.Conn, error)

type HttpClientBuilder interface {
	WithProxy() HttpClientBuilder
	WithoutProxy() HttpClientBuilder
	WithRedirects() HttpClientBuilder
	WithoutRedirects() HttpClientBuilder
	WithTimeout(timeout time.Duration) HttpClientBuilder
	WithDialFunc(dialFunc DialFunc) HttpClientBuilder
	WithInsecureSkipVerify(insecureSkipVerify bool) HttpClientBuilder
	WithClientCert(caCertPath string, clientCertPath string, clientKeyPath string) HttpClientBuilder
	Build(ctx context.Context) (*http.Client, error)
	BuildRoundTripper(ctx context.Context) (http.RoundTripper, error)
}

type httpClientBuilder struct {
	proxy              Proxy
	checkRedirect      CheckRedirect
	timeout            time.Duration
	dialFunc           DialFunc
	insecureSkipVerify bool
	caCertPath         string
	clientCertPath     string
	clientKeyPath      string
}

func (h *httpClientBuilder) WithClientCert(caCertPath string, clientCertPath string, clientKeyPath string) HttpClientBuilder {
	h.caCertPath = caCertPath
	h.clientCertPath = clientCertPath
	h.clientKeyPath = clientKeyPath
	return h
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

func (h *httpClientBuilder) BuildRoundTripper(ctx context.Context) (http.RoundTripper, error) {
	if glog.V(5) {
		glog.Infof("build http transport")
	}
	tlsClientConfig := &tls.Config{}
	if h.caCertPath != "" && h.clientCertPath != "" && h.clientKeyPath != "" {
		var err error
		tlsClientConfig, err = CreateTlsClientConfig(ctx, h.caCertPath, h.clientCertPath, h.clientKeyPath)
		if err != nil {
			return nil, errors.Wrapf(ctx, err, "create tls config failed")
		}
	}
	tlsClientConfig.InsecureSkipVerify = h.insecureSkipVerify
	return &http.Transport{
		Proxy:           h.proxy,
		DialContext:     h.BuildDialFunc(),
		TLSClientConfig: tlsClientConfig,
	}, nil
}

func (h *httpClientBuilder) Build(ctx context.Context) (*http.Client, error) {
	if glog.V(5) {
		glog.Infof("build http client")
	}
	roundTripper, err := h.BuildRoundTripper(ctx)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "build roundTripper failed")
	}

	return &http.Client{
		Transport:     roundTripper,
		CheckRedirect: h.checkRedirect,
	}, nil
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
			return stderrors.New("stopped after 10 redirects")
		}
		return nil
	}
	return h
}

func (h *httpClientBuilder) WithoutRedirects() HttpClientBuilder {
	h.checkRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 1 {
			return stderrors.New("redirects")
		}
		return nil
	}
	return h
}

func (h *httpClientBuilder) WithInsecureSkipVerify(insecureSkipVerify bool) HttpClientBuilder {
	h.insecureSkipVerify = insecureSkipVerify
	return nil
}
