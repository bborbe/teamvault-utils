package main

import (
	"context"
	"os"

	"github.com/bborbe/errors"
	libsentry "github.com/bborbe/sentry"
	"github.com/bborbe/service"

	"github.com/bborbe/teamvault-utils/v4"
	"github.com/bborbe/teamvault-utils/v4/factory"
)

func main() {
	app := &application{}
	os.Exit(service.Main(context.Background(), app, &app.SentryDSN, &app.SentryProxy))
}

type application struct {
	SentryDSN           string `required:"false" arg:"sentry-dsn" env:"SENTRY_DSN" usage:"SentryDSN" display:"length"`
	SentryProxy         string `required:"false" arg:"sentry-proxy" env:"SENTRY_PROXY" usage:"Sentry Proxy"`
	TeamvaultUrl        string `required:"false" arg:"teamvault-url" env:"TEAMVAULT_URL" usage:"teamvault url"`
	TeamvaultUser       string `required:"false" arg:"teamvault-user" env:"TEAMVAULT_USER" usage:"teamvault user"`
	TeamvaultPass       string `required:"false" arg:"teamvault-pass" env:"TEAMVAULT_PASS" usage:"teamvault password" display:"length"`
	TeamvaultConfigPath string `required:"false" arg:"teamvault-config" env:"TEAMVAULT_CONFIG" usage:"teamvault config"`
	SourceDirectory     string `required:"true" arg:"source-dir" env:"SOURCE_DIR" usage:"source directory"`
	TargetDirectory     string `required:"true" arg:"target-dir" env:"TARGET_DIR" usage:"target directory"`
	Staging             bool   `required:"false" arg:"staging" env:"STAGING" usage:"staging status" default:"false"`
	Cache               bool   `required:"false" arg:"cache" env:"CACHE" usage:"enable teamvault secret cache" default:"false"`
}

func (a *application) Run(ctx context.Context, sentryClient libsentry.Client) error {
	httpClient, err := factory.CreateHttpClient(ctx)
	if err != nil {
		return errors.Wrapf(ctx, err, "create httpClient failed")
	}

	teamvaultConnector, err := factory.CreateConnectorWithConfig(
		ctx,
		httpClient,
		teamvault.TeamvaultConfigPath(a.TeamvaultConfigPath),
		teamvault.Url(a.TeamvaultUrl),
		teamvault.User(a.TeamvaultUser),
		teamvault.Password(a.TeamvaultPass),
		teamvault.Staging(a.Staging),
		a.Cache,
	)
	if err != nil {
		return errors.Wrapf(ctx, err, "create connector failed")
	}
	sourceDirectory := teamvault.SourceDirectory(a.SourceDirectory)
	targetDirectory := teamvault.TargetDirectory(a.TargetDirectory)
	configParser := teamvault.NewConfigParser(teamvaultConnector)
	manifestsGenerator := teamvault.NewConfigGenerator(configParser)
	if err := manifestsGenerator.Generate(ctx, sourceDirectory, targetDirectory); err != nil {
		return errors.Wrapf(ctx, err, "generate failed")
	}
	return nil
}
