package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"

	"github.com/bborbe/http/client_builder"
	"github.com/bborbe/teamvault-utils"
	"github.com/bborbe/teamvault-utils/connector"
	"github.com/golang/glog"
)

var (
	teamvaultUrlPtr        = flag.String("teamvault-url", "", "teamvault url")
	teamvaultUserPtr       = flag.String("teamvault-user", "", "teamvault user")
	teamvaultPassPtr       = flag.String("teamvault-pass", "", "teamvault password")
	teamvaultConfigPathPtr = flag.String("teamvault-config", "", "teamvault config")
	stagingPtr             = flag.Bool("staging", false, "staging status")
	teamvaultKeyPtr        = flag.String("teamvault-key", "", "teamvault key")
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	err := do()
	if err != nil {
		glog.Exit(err)
	}
}

func do() error {
	teamvaultUrl := teamvault.TeamvaultUrl(*teamvaultUrlPtr)
	teamvaultUser := teamvault.TeamvaultUser(*teamvaultUserPtr)
	teamvaultPassword := teamvault.TeamvaultPassword(*teamvaultPassPtr)
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
	httpClient := client_builder.New().WithTimeout(5 * time.Second).Build()
	var teamvaultConnector teamvault.Connector
	if !staging {
		teamvaultConnector = connector.New(httpClient.Do, teamvaultUrl, teamvaultUser, teamvaultPassword)
	} else {
		teamvaultConnector = connector.NewDummy()
	}
	result, err := teamvaultConnector.Password(teamvault.TeamvaultKey(*teamvaultKeyPtr))
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", result)
	return nil
}
