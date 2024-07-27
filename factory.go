package teamvault

import (
	"net/http"
	"time"

	libhttp "github.com/bborbe/http"
	"github.com/golang/glog"
)

func CreateConnectorWithConfig(
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
		apiURL,
		apiUser,
		apiPassword,
		staging,
		cacheEnabled,
	), nil
}

func CreateConnector(
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
			CreateRemoteConnector(apiURL, apiUser, apiPassword),
		)
	}
	return CreateRemoteConnector(apiURL, apiUser, apiPassword)
}

func CreateRemoteConnector(
	apiURL Url,
	apiUser User,
	apiPassword Password,
) Connector {
	return NewRemoteConnector(
		CreateHttpClient(),
		apiURL,
		apiUser,
		apiPassword,
	)
}

func CreateHttpClient() *http.Client {
	return libhttp.NewClientBuilder().WithTimeout(5 * time.Second).Build()
}
