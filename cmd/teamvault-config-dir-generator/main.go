package main

import (
	"context"
	"flag"
	"runtime"

	"github.com/bborbe/errors"
	"github.com/golang/glog"

	"github.com/bborbe/teamvault-utils"
)

var (
	teamvaultUrlPtr        = flag.String("teamvault-url", "", "teamvault url")
	teamvaultUserPtr       = flag.String("teamvault-user", "", "teamvault user")
	teamvaultPassPtr       = flag.String("teamvault-pass", "", "teamvault password")
	teamvaultConfigPathPtr = flag.String("teamvault-config", "", "teamvault config")
	sourceDirectoryPtr     = flag.String("source-dir", "", "source directory")
	targetDirectoryPtr     = flag.String("target-dir", "", "target directory")
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
	httpClient, err := teamvault.CreateHttpClient(ctx)
	if err != nil {
		return errors.Wrapf(ctx, err, "create httpClient failed")
	}

	teamvaultConnector, err := teamvault.CreateConnectorWithConfig(
		httpClient,
		teamvault.TeamvaultConfigPath(*teamvaultConfigPathPtr),
		teamvault.Url(*teamvaultUrlPtr),
		teamvault.User(*teamvaultUserPtr),
		teamvault.Password(*teamvaultPassPtr),
		teamvault.Staging(*stagingPtr),
		*cachePtr,
	)
	if err != nil {
		return errors.Wrapf(ctx, err, "create connector failed")
	}
	sourceDirectory := teamvault.SourceDirectory(*sourceDirectoryPtr)
	targetDirectory := teamvault.TargetDirectory(*targetDirectoryPtr)
	configParser := teamvault.NewConfigParser(teamvaultConnector)
	manifestsGenerator := teamvault.NewConfigGenerator(configParser)
	if err := manifestsGenerator.Generate(ctx, sourceDirectory, targetDirectory); err != nil {
		return err
	}
	return nil
}
