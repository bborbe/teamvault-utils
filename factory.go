package teamvault

import (
	"context"
	"net/http"
	"time"

	libhttp "github.com/bborbe/http"
	"github.com/golang/glog"
)

func CreateConnectorWithConfig(
	httpClient *http.Client,
	configPath TeamvaultConfigPath,
	apiURL Url,
	apiUser User,
	apiPassword Password,
	staging Staging,
	cacheEnabled bool,
) (Connector, error) {
	if configPath.Exists() {
		config, err := configPath.Parse()
		if err != nil {
			glog.V(2).Infof("parse teamvault config failed: %v", err)
			return nil, err
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
	apiURL Url,
	apiUser User,
	apiPassword Password,
	staging Staging,
	cacheEnabled bool,
) Connector {
	if staging {
		return NewDummyConnector()
	}
	if cacheEnabled {
		return NewDiskFallbackConnector(
			CreateRemoteConnector(httpClient, apiURL, apiUser, apiPassword),
		)
	}
	return CreateRemoteConnector(httpClient, apiURL, apiUser, apiPassword)
}

func CreateRemoteConnector(
	httpClient *http.Client,
	apiURL Url,
	apiUser User,
	apiPassword Password,
) Connector {
	return NewRemoteConnector(
		httpClient,
		apiURL,
		apiUser,
		apiPassword,
	)
}

func CreateHttpClient(ctx context.Context) (*http.Client, error) {
	return libhttp.NewClientBuilder().WithTimeout(5 * time.Second).Build(ctx)
}
