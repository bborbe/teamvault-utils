package main

import (
	"flag"
	"github.com/golang/glog"
	"runtime"
)

const (
	PARAMETER_TEAMVAULT_URL    = "teamvault-url"
	PARAMETER_TEAMVAULT_USER   = "teamvault-user"
	PARAMETER_TEAMVAULT_PASS   = "teamvault-pass"
	PARAMETER_SOURCE_DIRECTORY = "source-dir"
	PARAMETER_TARGET_DIRECTORY = "target-dir"
)

var (
	teamvaultUrlPtr    = flag.String(PARAMETER_TEAMVAULT_URL, "", "teamvault url")
	teamvaultUserPtr   = flag.String(PARAMETER_TEAMVAULT_USER, "", "teamvault user")
	teamvaultPassPtr   = flag.String(PARAMETER_TEAMVAULT_PASS, "", "teamvault password")
	sourceDirectoryPtr = flag.String(PARAMETER_SOURCE_DIRECTORY, "", "source directory")
	targetDirectoryPtr = flag.String(PARAMETER_TARGET_DIRECTORY, "", "target directory")
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

type variableName string
type teamvaultKey string
type sourceDirectory string
type targetDirectory string

func do() error {
	var data map[variableName]teamvaultKey
	data["news_hipchat_password"] = "m7yyRv"

	//sourceDirectory := sourceDirectory(*sourceDirectoryPtr)
	//targetDirectory := targetDirectory(*targetDirectoryPtr)

	return nil
}
