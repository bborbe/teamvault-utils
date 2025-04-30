package factory

import (
	"context"
	"github.com/bborbe/errors"
	"net/http"
	"time"

	libhttp "github.com/bborbe/http"
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
	), nil
}

func CreateConnector(
	httpClient *http.Client,
	apiURL teamvault.Url,
	apiUser teamvault.User,
	apiPassword teamvault.Password,
	staging teamvault.Staging,
	cacheEnabled bool,
) teamvault.Connector {
	if staging {
		return teamvault.NewDummyConnector()
	}
	if cacheEnabled {
		return teamvault.NewDiskFallbackConnector(
			CreateRemoteConnector(httpClient, apiURL, apiUser, apiPassword),
		)
	}
	return CreateRemoteConnector(httpClient, apiURL, apiUser, apiPassword)
}

func CreateRemoteConnector(
	httpClient *http.Client,
	apiURL teamvault.Url,
	apiUser teamvault.User,
	apiPassword teamvault.Password,
) teamvault.Connector {
	return teamvault.NewRemoteConnector(
		httpClient,
		apiURL,
		apiUser,
		apiPassword,
	)
}

func CreateHttpClient(ctx context.Context) (*http.Client, error) {
	return libhttp.NewClientBuilder().WithTimeout(5 * time.Second).Build(ctx)
}
