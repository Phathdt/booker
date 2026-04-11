// @title          Booker CEX API
// @version        1.0
// @description    Centralized Exchange demo — token trading platform
// @host           localhost
// @BasePath       /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}"
package main

import (
	"log"
	"os"

	mycli "booker/cli"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "booker",
		Usage: "CEX Booking Engine — Centralized Exchange Demo",
		Commands: []*cli.Command{
			{
				Name:    "users-svc",
				Aliases: []string{"us"},
				Usage:   "Start the users/auth service (gRPC + REST)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Value:   "config.yml",
						Usage:   "Configuration file path",
					},
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Value:   50051,
						Usage:   "gRPC listen port",
					},
					&cli.IntFlag{
						Name:  "http-port",
						Value: 8081,
						Usage: "REST API port (0 = disabled)",
					},
				},
				Action: mycli.RunUsersSvc,
			},
			{
				Name:    "wallet-svc",
				Aliases: []string{"ws"},
				Usage:   "Start the wallet service (gRPC + REST)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Value:   "config.yml",
						Usage:   "Configuration file path",
					},
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Value:   50052,
						Usage:   "gRPC listen port",
					},
					&cli.IntFlag{
						Name:  "http-port",
						Value: 8082,
						Usage: "REST API port (0 = disabled)",
					},
				},
				Action: mycli.RunWalletSvc,
			},
			{
				Name:    "migrate",
				Aliases: []string{"m"},
				Usage:   "Run database migrations",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Value:   "config.yml",
						Usage:   "Configuration file path",
					},
				},
				Subcommands: []*cli.Command{
					{
						Name:   "up",
						Usage:  "Run all pending migrations",
						Action: mycli.MigrateUp,
					},
					{
						Name:   "down",
						Usage:  "Rollback the last migration",
						Action: mycli.MigrateDown,
					},
					{
						Name:   "status",
						Usage:  "Show migration status",
						Action: mycli.MigrateStatus,
					},
				},
			},
		},
		Action: func(c *cli.Context) error {
			return cli.ShowAppHelp(c)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
