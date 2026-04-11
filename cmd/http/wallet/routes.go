package wallet

import (
	"booker/modules/users/domain/interfaces"
	walletInterfaces "booker/modules/wallet/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes sets up wallet HTTP routes on the Fiber app.
func RegisterRoutes(app *fiber.App, walletSvc walletInterfaces.WalletService, tokenSvc interfaces.TokenService) {
	w := app.Group("/api/v1/wallet", httpserver.AuthMiddleware(tokenSvc))

	w.Get("/", GetBalances(walletSvc))
	w.Get("/:asset_id", GetBalance(walletSvc))
	w.Post("/deposit", Deposit(walletSvc))
	w.Post("/withdraw", Withdraw(walletSvc))
}
