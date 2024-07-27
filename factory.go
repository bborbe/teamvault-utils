package teamvault

import (
	"net/http"
	"time"

	libhttp "github.com/bborbe/http"
	"github.com/golang/glog"
)

func CreateConnector(
	teamvaultConfigPath TeamvaultConfigPath,
	teamvaultUrl Url,
	teamvaultUser User,
	teamvaultPassword Password,
	staging Staging,
	cache bool,
) (Connector, error) {
	if staging {
		return NewDummyConnector(), nil
	}
	if teamvaultConfigPath.Exists() {
		teamvaultConfig, err := teamvaultConfigPath.Parse()
		if err != nil {
			glog.V(2).Infof("parse teamvault config failed: %v", err)
			return nil, err
		}
		teamvaultUrl = teamvaultConfig.Url
		teamvaultUser = teamvaultConfig.User
		teamvaultPassword = teamvaultConfig.Password
	}
	var teamvaultConnector Connector
	teamvaultConnector = NewRemoteConnector(
		CreateHttpClient(),
		teamvaultUrl,
		teamvaultUser,
		teamvaultPassword,
	)
	if cache {
		teamvaultConnector = NewDiskFallbackConnector(teamvaultConnector)
	}
	return teamvaultConnector, nil
}

func CreateHttpClient() *http.Client {
	return libhttp.NewClientBuilder().WithTimeout(5 * time.Second).Build()
}
