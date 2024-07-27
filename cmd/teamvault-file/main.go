package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"

	"github.com/bborbe/errors"
	"github.com/golang/glog"

	"github.com/bborbe/teamvault-utils"
)

var (
	teamvaultURLPtr        = flag.String("teamvault-url", "", "teamvault url")
	teamvaultUserPtr       = flag.String("teamvault-user", "", "teamvault user")
	teamvaultPassPtr       = flag.String("teamvault-pass", "", "teamvault password")
	teamvaultConfigPathPtr = flag.String("teamvault-config", "", "teamvault config")
	stagingPtr             = flag.Bool("staging", false, "staging status")
	teamvaultKeyPtr        = flag.String("teamvault-key", "", "teamvault key")
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
	teamvaultConnector, err := teamvault.CreateConnectorWithConfig(
		teamvault.TeamvaultConfigPath(*teamvaultConfigPathPtr),
		teamvault.Url(*teamvaultURLPtr),
		teamvault.User(*teamvaultUserPtr),
		teamvault.Password(*teamvaultPassPtr),
		teamvault.Staging(*stagingPtr),
		*cachePtr,
	)
	if err != nil {
		return errors.Wrapf(ctx, err, "create connector failed")
	}

	result, err := teamvaultConnector.File(ctx, teamvault.Key(*teamvaultKeyPtr))
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", result)
	return nil
}
