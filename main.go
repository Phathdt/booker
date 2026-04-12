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
				Name:    "order-svc",
				Aliases: []string{"os"},
				Usage:   "Start the order service (gRPC + REST)",
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
						Value:   50053,
						Usage:   "gRPC listen port",
					},
					&cli.IntFlag{
						Name:  "http-port",
						Value: 8083,
						Usage: "REST API port (0 = disabled)",
					},
				},
				Action: mycli.RunOrderSvc,
			},
			{
				Name:    "matching-svc",
				Aliases: []string{"ms"},
				Usage:   "Start the matching engine service (gRPC only)",
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
						Value:   50054,
						Usage:   "gRPC listen port",
					},
				},
				Action: mycli.RunMatchingSvc,
			},
			{
				Name:    "market-svc",
				Aliases: []string{"mk"},
				Usage:   "Start the market data service (REST + WebSocket)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Value:   "config.yml",
						Usage:   "Configuration file path",
					},
					&cli.IntFlag{
						Name:  "http-port",
						Value: 8085,
						Usage: "REST + WS API port",
					},
				},
				Action: mycli.RunMarketSvc,
			},
			{
				Name:    "notification-svc",
				Aliases: []string{"ns"},
				Usage:   "Start the notification service (REST + WebSocket)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Value:   "config.yml",
						Usage:   "Configuration file path",
					},
					&cli.IntFlag{
						Name:  "http-port",
						Value: 8086,
						Usage: "REST + WebSocket port",
					},
				},
				Action: mycli.RunNotificationSvc,
			},
			{
				Name:    "swagger-svc",
				Aliases: []string{"sw"},
				Usage:   "Start the Swagger UI service (REST only)",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "http-port",
						Value: 8090,
						Usage: "REST API port for Swagger UI",
					},
				},
				Action: mycli.RunSwaggerSvc,
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
