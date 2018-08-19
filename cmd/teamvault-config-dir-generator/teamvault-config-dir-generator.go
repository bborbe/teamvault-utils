package main

import (
	"flag"
	"runtime"
	"time"

	"github.com/bborbe/http/client_builder"
	"github.com/bborbe/teamvault-utils"
	"github.com/bborbe/teamvault-utils/connector"
	"github.com/bborbe/teamvault-utils/generator"
	"github.com/bborbe/teamvault-utils/parser"
	"github.com/golang/glog"
)

var (
	teamvaultUrlPtr        = flag.String("teamvault-url", "", "teamvault url")
	teamvaultUserPtr       = flag.String("teamvault-user", "", "teamvault user")
	teamvaultPassPtr       = flag.String("teamvault-pass", "", "teamvault password")
	teamvaultConfigPathPtr = flag.String("teamvault-config", "", "teamvault config")
	sourceDirectoryPtr     = flag.String("source-dir", "", "source directory")
	targetDirectoryPtr     = flag.String("target-dir", "", "target directory")
	stagingPtr             = flag.Bool("staging", false, "staging status")
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
	sourceDirectory := teamvault.SourceDirectory(*sourceDirectoryPtr)
	targetDirectory := teamvault.TargetDirectory(*targetDirectoryPtr)
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
	configParser := parser.New(teamvaultConnector)
	manifestsGenerator := generator.New(configParser)
	if err := manifestsGenerator.Generate(sourceDirectory, targetDirectory); err != nil {
		return err
	}
	return nil
}