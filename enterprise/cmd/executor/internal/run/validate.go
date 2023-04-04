package run

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sourcegraph/log"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Validate(cliCtx *cli.Context, runner util.CmdRunner, logger log.Logger, config *config.Config) error {
	// First, validate the config is valid.
	if err := config.Validate(); err != nil {
		return err
	}

	// Then, validate all tools that are required are installed.
	if err := util.ValidateRequiredTools(runner, config.UseFirecracker); err != nil {
		return err
	}

	// Validate git is of the right version.
	if err := util.ValidateGitVersion(cliCtx.Context, runner); err != nil {
		return err
	}

	telemetryOptions := newQueueTelemetryOptions(cliCtx.Context, runner, config.UseFirecracker, logger)
	copts := queueOptions(config, telemetryOptions)
	client, err := apiclient.NewBaseClient(copts.BaseClientOptions)
	if err != nil {
		return err
	}

	// Validate frontend access token returns status 200.
	if err = validateAuthorizationToken(cliCtx.Context, copts.BaseClientOptions); err != nil {
		return err
	}

	// Validate src-cli is of a good version, rely on the connected instance to tell
	// us what "good" means.
	if err = util.ValidateSrcCLIVersion(cliCtx.Context, runner, client, copts.BaseClientOptions.EndpointOptions); err != nil {
		if errors.Is(err, util.ErrSrcPatchBehind) {
			// This is ok. The patch just doesn't match but still works.
			logger.Warn("A newer patch release version of src-cli is available, consider running executor install src-cli to upgrade", log.Error(err))
		} else {
			return err
		}
	}

	if config.UseFirecracker {
		// Validate ignite is installed.
		if err = util.ValidateIgniteInstalled(cliCtx.Context, runner); err != nil {
			return err
		}
		// Validate all required CNI plugins are installed.
		if err = util.ValidateCNIInstalled(); err != nil {
			return err
		}

		// TODO: Validate ignite images are pulled and imported. Sadly, the
		// output of ignite is not very parser friendly.
	}

	fmt.Print("All checks passed!\n")

	return nil
}

var authorizationFailedErr = errors.New("failed to authorize with frontend, is executors.accessToken set correctly in the site-config?")

func validateAuthorizationToken(ctx context.Context, options apiclient.BaseClientOptions) error {
	options.EndpointOptions.PathPrefix = ""
	client, err := apiclient.NewBaseClient(options)
	if err != nil {
		return err
	}

	req, err := client.NewJSONRequest(http.MethodGet, ".executors/test/auth", nil)
	if err != nil {
		return err
	}

	if err = client.DoAndDrop(ctx, req); err != nil {
		var unexpectedStatusCodeError *apiclient.UnexpectedStatusCodeErr
		if errors.As(err, &unexpectedStatusCodeError) && (unexpectedStatusCodeError.StatusCode == http.StatusUnauthorized) {
			return authorizationFailedErr
		} else {
			return errors.Wrap(err, "failed to validate authorization token")
		}
	}

	return nil
}
