package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"runtime"
	"time"

	libhttp "github.com/bborbe/http"
	"github.com/golang/glog"

	"github.com/bborbe/teamvault-utils"
)

var (
	teamvaultUrlPtr        = flag.String("teamvault-url", "", "teamvault url")
	teamvaultUserPtr       = flag.String("teamvault-user", "", "teamvault user")
	teamvaultPassPtr       = flag.String("teamvault-pass", "", "teamvault password")
	teamvaultConfigPathPtr = flag.String("teamvault-config", "", "teamvault config")
	stagingPtr             = flag.Bool("staging", false, "staging status")
	cachePtr               = flag.Bool("cache", false, "enable teamvault secret cache")
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	err := do(context.Background())
	if err != nil {
		glog.Exit(err)
	}
}

func do(ctx context.Context) error {
	teamvaultUrl := teamvault.Url(*teamvaultUrlPtr)
	teamvaultUser := teamvault.User(*teamvaultUserPtr)
	teamvaultPassword := teamvault.Password(*teamvaultPassPtr)
	teamvaultConfigPath := teamvault.TeamvaultConfigPath(*teamvaultConfigPathPtr)
	staging := teamvault.Staging(*stagingPtr)
	if teamvaultConfigPath.Exists() {
		teamvaultConfig, err := teamvaultConfigPath.Parse()
		if err != nil {
			glog.V(2).Infof("parse teamvault config failed: %v", err)
			return err
		}
		teamvaultUrl = teamvaultConfig.Url
		teamvaultUser = teamvaultConfig.User
		teamvaultPassword = teamvaultConfig.Password
	}
	httpClient := libhttp.NewClientBuilder().WithTimeout(5 * time.Second).Build()
	var teamvaultConnector teamvault.Connector
	if !staging {
		teamvaultConnector = teamvault.NewRemoteConnector(httpClient, teamvaultUrl, teamvaultUser, teamvaultPassword)
		if *cachePtr {
			teamvaultConnector = teamvault.NewDiskFallbackConnector(teamvaultConnector)
		}
	} else {
		teamvaultConnector = teamvault.NewDummyConnector()
	}
	configParser := teamvault.NewParser(teamvaultConnector)
	content, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	output, err := configParser.Parse(ctx, content)
	if err != nil {
		return err
	}
	if _, err := os.Stdout.Write(output); err != nil {
		return err
	}
	return nil
}
