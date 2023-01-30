package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/containers/common/pkg/retry"
	"github.com/containers/image/v5/docker"
	"github.com/spf13/cobra"
)

type reposListOptions struct {
	global       *globalOptions
	image        *imageOptions
	retryOpts    *retry.Options
	searchFilter string
	limit        int
}

func reposListCmd(global *globalOptions) *cobra.Command {
	sharedFlags, sharedOpts := sharedImageFlags()
	imageFlags, imageOpts := dockerImageFlags(global, sharedOpts, nil, "", "")
	retryFlags, retryOpts := retryFlags()

	opts := reposListOptions{
		global:    global,
		image:     imageOpts,
		retryOpts: retryOpts,
	}
	cmd := &cobra.Command{
		Use:     "list-repos [command options] REGISTRY FILTER",
		Short:   "List repositories of a container registry",
		Long:    "List repositories of a container registry on a specified server.",
		RunE:    commandAction(opts.run),
		Example: `skopeo list-repos quay.io repo`,
	}
	adjustUsage(cmd)
	flags := cmd.Flags()

	flags.StringVar(&opts.searchFilter, "search", "", "String used to search the registry")
	flags.IntVar(&opts.limit, "limit", 100, "number of elements returned")

	flags.AddFlagSet(&sharedFlags)
	flags.AddFlagSet(&imageFlags)
	flags.AddFlagSet(&retryFlags)

	return cmd
}

func (opts *reposListOptions) run(args []string, stdout io.Writer) error {
	ctx, cancel := opts.global.commandTimeoutContext()
	defer cancel()

	if len(args) != 1 {
		return errorShouldDisplayUsage{errors.New("Exactly one non-option argument expected")}
	}

	sys, err := opts.image.newSystemContext()
	if err != nil {
		return err
	}

	outputData, err := docker.SearchRegistry(ctx, sys, args[0], opts.searchFilter, opts.limit)
	if err != nil {
		return err
	}

	out, err := json.MarshalIndent(outputData, "", "    ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "%s\n", string(out))
	return err
}
