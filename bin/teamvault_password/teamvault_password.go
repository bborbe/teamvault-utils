package main

import (
	"flag"
	"runtime"
	"time"

	"github.com/bborbe/http/client_builder"
	"github.com/bborbe/teamvault_utils/connector"
	"github.com/bborbe/teamvault_utils/model"
	"github.com/golang/glog"
	"fmt"
)

var (
	teamvaultUrlPtr        = flag.String(model.PARAMETER_TEAMVAULT_URL, "", "teamvault url")
	teamvaultUserPtr       = flag.String(model.PARAMETER_TEAMVAULT_USER, "", "teamvault user")
	teamvaultPassPtr       = flag.String(model.PARAMETER_TEAMVAULT_PASS, "", "teamvault password")
	teamvaultConfigPathPtr = flag.String(model.PARAMETER_TEAMVAULT_CONFIG, "", "teamvault config")
	stagingPtr             = flag.Bool(model.PARAMETER_STAGING, false, "staging status")
	teamvaultKeyPtr        = flag.String(model.PARAMETER_TEAMVAULT_KEY, "", "teamvault key")
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
	teamvaultUrl := model.TeamvaultUrl(*teamvaultUrlPtr)
	teamvaultUser := model.TeamvaultUser(*teamvaultUserPtr)
	teamvaultPassword := model.TeamvaultPassword(*teamvaultPassPtr)
	teamvaultConfigPath := model.TeamvaultConfigPath(*teamvaultConfigPathPtr)
	staging := model.Staging(*stagingPtr)
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
	var teamvaultConnector connector.Connector
	if !staging {
		teamvaultConnector = connector.New(httpClient.Do, teamvaultUrl, teamvaultUser, teamvaultPassword)
	} else {
		teamvaultConnector = connector.NewDummy()
	}
	result, err := teamvaultConnector.Password(model.TeamvaultKey(*teamvaultKeyPtr))
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", result)
	return nil
}
