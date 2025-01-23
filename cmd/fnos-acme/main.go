/*
 * Copyright 2024 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
)

const (
	flgDataDir      = "data-dir"
	flgFnosAddress  = "fnos-address"
	flgFnosUsername = "fnos-username"
	flgFnosPassword = "fnos-password"
	flgDomains      = "domains"
	flgEmail        = "email"
	flgDnsProvider  = "dns-provider"
	flgDnsResolvers = "dns-resolvers"
	flgDebug        = "debug"
)

func main() {
	root := &cli.Command{
		Name: "fnos-acme",
		Commands: []*cli.Command{
			commandRun(),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    flgDataDir,
				Value:   "/app/fnos-acme",
				Usage:   "data dir",
				Sources: cli.EnvVars("DATA_DIR"),
			},
			&cli.StringFlag{
				Name:    flgFnosAddress,
				Value:   "",
				Usage:   "FNOS address",
				Sources: cli.EnvVars("FNOS_ADDRESS"),
			},
			&cli.StringFlag{
				Name:    flgFnosUsername,
				Value:   "",
				Usage:   "FNOS username",
				Sources: cli.EnvVars("FNOS_USERNAME"),
			},
			&cli.StringFlag{
				Name:    flgFnosPassword,
				Value:   "",
				Usage:   "FNOS password",
				Sources: cli.EnvVars("FNOS_PASSWORD"),
			},
			&cli.StringSliceFlag{
				Name:    flgDomains,
				Value:   []string{},
				Usage:   "main domain",
				Sources: cli.EnvVars("DOMAINS"),
			},
			&cli.StringFlag{
				Name:    flgEmail,
				Value:   "",
				Usage:   "email",
				Sources: cli.EnvVars("EMAIL"),
			},
			&cli.StringFlag{
				Name:    flgDnsProvider,
				Value:   "",
				Usage:   "dns provider",
				Sources: cli.EnvVars("DNS_PROVIDER"),
			},
			&cli.StringSliceFlag{
				Name:    flgDnsResolvers,
				Value:   []string{},
				Usage:   "dns resolvers",
				Sources: cli.EnvVars("DNS_RESOLVERS"),
			},
			&cli.BoolFlag{
				Name:    flgDebug,
				Value:   false,
				Usage:   "debug mode",
				Sources: cli.EnvVars("DEBUG"),
			},
		},
	}

	if err := root.Run(context.Background(), os.Args); err != nil {
		slog.Error("run failed", "err", err)
		os.Exit(1)
	}
}
