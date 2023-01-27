package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/containers/common/pkg/auth"
	commonFlag "github.com/containers/common/pkg/flag"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"
	"github.com/spf13/cobra"
)

type reposListOptions struct {
	global    *globalOptions
	loginOpts auth.LoginOptions
	tlsVerify commonFlag.OptionalBool
}

func reposListCmd(global *globalOptions) *cobra.Command {
	opts := reposListOptions{
		global: global,
	}
	cmd := &cobra.Command{
		Use:     "list-repos [command options] REGISTRY",
		Short:   "List repositories of a container registry",
		Long:    "List repositories of a container registry on a specified server.",
		RunE:    commandAction(opts.run),
		Example: `skopeo list-repos quay.io repo`,
	}
	adjustUsage(cmd)
	flags := cmd.Flags()
	commonFlag.OptionalBoolFlag(flags, &opts.tlsVerify, "tls-verify", "require HTTPS and verify certificates when accessing the registry")
	flags.AddFlagSet(auth.GetLoginFlags(&opts.loginOpts))
	return cmd
}

func (opts *reposListOptions) run(args []string, stdout io.Writer) error {
	ctx, cancel := opts.global.commandTimeoutContext()
	defer cancel()

	if len(args) != 2 {
		return errorShouldDisplayUsage{errors.New("Exactly two non-option argument expected")}
	}

	opts.loginOpts.Stdout = stdout
	opts.loginOpts.Stdin = os.Stdin
	opts.loginOpts.AcceptRepositories = true
	sys := opts.global.newSystemContext()
	if opts.tlsVerify.Present() {
		sys.DockerInsecureSkipTLSVerify = types.NewOptionalBool(!opts.tlsVerify.Value())
	}
	outputData, err := docker.SearchRegistry(ctx, sys, args[0], args[1], 2)
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
