// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package factory

import (
	"context"
	"net/http"
	"time"

	"github.com/bborbe/errors"
	libhttp "github.com/bborbe/http"
	libtime "github.com/bborbe/time"

	"github.com/bborbe/teamvault-utils/v4"
)

func CreateConnectorWithConfig(
	ctx context.Context,
	httpClient *http.Client,
	configPath teamvault.TeamvaultConfigPath,
	apiURL teamvault.Url,
	apiUser teamvault.User,
	apiPassword teamvault.Password,
	staging teamvault.Staging,
	cacheEnabled bool,
	currentDateTime libtime.CurrentDateTime,
) (teamvault.Connector, error) {
	if configPath.Exists() {
		config, err := configPath.Parse()
		if err != nil {
			return nil, errors.Wrapf(ctx, err, "parse teamvault config failed")
		}
		apiURL = config.Url
		apiUser = config.User
		apiPassword = config.Password
		cacheEnabled = config.CacheEnabled
	}
	return CreateConnector(
		httpClient,
		apiURL,
		apiUser,
		apiPassword,
		staging,
		cacheEnabled,
		currentDateTime,
	), nil
}

func CreateConnector(
	httpClient *http.Client,
	apiURL teamvault.Url,
	apiUser teamvault.User,
	apiPassword teamvault.Password,
	staging teamvault.Staging,
	cacheEnabled bool,
	currentDateTime libtime.CurrentDateTime,
) teamvault.Connector {
	if staging {
		return teamvault.NewDummyConnector()
	}
	if cacheEnabled {
		return teamvault.NewDiskFallbackConnector(
			CreateRemoteConnector(httpClient, apiURL, apiUser, apiPassword, currentDateTime),
		)
	}
	return CreateRemoteConnector(httpClient, apiURL, apiUser, apiPassword, currentDateTime)
}

func CreateRemoteConnector(
	httpClient *http.Client,
	apiURL teamvault.Url,
	apiUser teamvault.User,
	apiPassword teamvault.Password,
	currentDateTime libtime.CurrentDateTime,
) teamvault.Connector {
	return teamvault.NewRemoteConnector(
		httpClient,
		apiURL,
		apiUser,
		apiPassword,
		currentDateTime,
	)
}

func CreateHttpClient(ctx context.Context) (*http.Client, error) {
	return libhttp.NewClientBuilder().WithTimeout(5 * time.Second).Build(ctx)
}
