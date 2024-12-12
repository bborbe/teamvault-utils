// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/bborbe/errors"
)

func CreateDefaultRoundTripper() RoundTripper {
	return createDefaultRoundTripper(nil)
}

func CreateDefaultRoundTripperTls(ctx context.Context, caCertPath string, clientCertPath string, clientKeyPath string) (RoundTripper, error) {
	tlsClientConfig, err := CreateTlsClientConfig(ctx, caCertPath, clientCertPath, clientKeyPath)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "create tls client config failed")
	}
	return createDefaultRoundTripper(tlsClientConfig), nil
}

func CreateTlsClientConfig(ctx context.Context, caCertPath string, clientCertPath string, clientKeyPath string) (*tls.Config, error) {
	// Load the client certificate and private key
	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "load client certificate and key failed")
	}

	// Load the CA certificate to verify the server
	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "read CA certificate failed")
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCertPEM); !ok {
		return nil, errors.Wrapf(ctx, err, "append CA certificate to pool failed")
	}

	// Set up TLS configuration with the client certificate and CA
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false, // Ensures server certificate is verified
	}
	return tlsConfig, nil
}

func createDefaultRoundTripper(tlsClientConfig *tls.Config) RoundTripper {
	return NewRoundTripperRetry(
		NewRoundTripperLog(
			&http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: defaultTransportDialContext(&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}),
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				ResponseHeaderTimeout: 30 * time.Second,
				TLSClientConfig:       tlsClientConfig,
			},
		),
		5,
		time.Second,
	)
}

func defaultTransportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}
