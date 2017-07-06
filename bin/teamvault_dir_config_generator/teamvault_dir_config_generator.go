package main

import (
	"flag"
	"runtime"
	"time"

	"github.com/bborbe/http/client_builder"
	"github.com/bborbe/teamvault_utils/teamvault"
	"github.com/bborbe/teamvault_utils/generator"
	"github.com/bborbe/teamvault_utils/model"
	"github.com/golang/glog"
)

var (
	teamvaultUrlPtr = flag.String(model.PARAMETER_TEAMVAULT_URL, "", "teamvault url")
	teamvaultUserPtr = flag.String(model.PARAMETER_TEAMVAULT_USER, "", "teamvault user")
	teamvaultPassPtr = flag.String(model.PARAMETER_TEAMVAULT_PASS, "", "teamvault password")
	teamvaultConfigPathPtr = flag.String(model.PARAMETER_TEAMVAULT_CONFIG, "", "teamvault config")
	sourceDirectoryPtr = flag.String(model.PARAMETER_SOURCE_DIRECTORY, "", "source directory")
	targetDirectoryPtr = flag.String(model.PARAMETER_TARGET_DIRECTORY, "", "target directory")
	stagingPtr = flag.Bool(model.PARAMETER_STAGING, false, "staging status")
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
	sourceDirectory := model.SourceDirectory(*sourceDirectoryPtr)
	targetDirectory := model.TargetDirectory(*targetDirectoryPtr)
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
	if !staging {
		tv := teamvault.New(httpClient.Do, teamvaultUrl, teamvaultUser, teamvaultPassword)
		manifestsGenerator := generator.New(tv.User, tv.Password, tv.Url, tv.File)
		if err := manifestsGenerator.Generate(sourceDirectory, targetDirectory); err != nil {
			return err
		}
	} else {
		tv := teamvault.NewDummy()
		manifestsGenerator := generator.New(tv.User, tv.Password, tv.URL, tv.File)
		if err := manifestsGenerator.Generate(sourceDirectory, targetDirectory); err != nil {
			return err
		}
	}
	return nil
}
