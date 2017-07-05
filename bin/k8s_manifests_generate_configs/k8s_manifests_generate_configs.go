package main

import (
	"flag"
	"runtime"
	"time"

	"github.com/bborbe/http/client_builder"
	"github.com/bborbe/kubernetes_tools/manifests/dummy"
	"github.com/bborbe/kubernetes_tools/manifests/generator"
	"github.com/bborbe/kubernetes_tools/manifests/model"
	"github.com/bborbe/kubernetes_tools/manifests/teamvault"
	"github.com/golang/glog"
)

const (
	PARAMETER_TEAMVAULT_URL    = "teamvault-url"
	PARAMETER_TEAMVAULT_USER   = "teamvault-user"
	PARAMETER_TEAMVAULT_PASS   = "teamvault-pass"
	PARAMETER_TEAMVAULT_CONFIG = "teamvault-config"
	PARAMETER_SOURCE_DIRECTORY = "source-dir"
	PARAMETER_TARGET_DIRECTORY = "target-dir"
	PARAMETER_STAGING          = "staging"
)

var (
	teamvaultUrlPtr        = flag.String(PARAMETER_TEAMVAULT_URL, "", "teamvault url")
	teamvaultUserPtr       = flag.String(PARAMETER_TEAMVAULT_USER, "", "teamvault user")
	teamvaultPassPtr       = flag.String(PARAMETER_TEAMVAULT_PASS, "", "teamvault password")
	teamvaultConfigPathPtr = flag.String(PARAMETER_TEAMVAULT_CONFIG, "", "teamvault config")
	sourceDirectoryPtr     = flag.String(PARAMETER_SOURCE_DIRECTORY, "", "source directory")
	targetDirectoryPtr     = flag.String(PARAMETER_TARGET_DIRECTORY, "", "target directory")
	stagingPtr             = flag.Bool(PARAMETER_STAGING, false, "staging status")
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
		tv := dummy.New()
		manifestsGenerator := generator.New(tv.User, tv.Password, tv.URL, tv.File)
		if err := manifestsGenerator.Generate(sourceDirectory, targetDirectory); err != nil {
			return err
		}
	}

	return nil
}
