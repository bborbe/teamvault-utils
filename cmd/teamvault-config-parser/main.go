package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/bborbe/errors"
	"github.com/golang/glog"

	"github.com/bborbe/teamvault-utils/v4"
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

	configParser := teamvault.NewConfigParser(teamvaultConnector)
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
