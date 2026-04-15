package wallet

import (
	"booker/modules/users/domain/interfaces"
	walletDTO "booker/modules/wallet/application/dto"
	walletInterfaces "booker/modules/wallet/domain/interfaces"
	"booker/pkg/httpserver"

	"github.com/oaswrap/spec/adapter/fiberopenapi"
	"github.com/oaswrap/spec/option"
)

// RegisterRoutes sets up wallet HTTP routes.
func RegisterRoutes(r fiberopenapi.Router, walletSvc walletInterfaces.WalletService, tokenSvc interfaces.TokenService) {
	w := r.Group("/api/v1/wallet", httpserver.AuthMiddleware(tokenSvc)).With(
		option.GroupSecurity("BearerAuth"),
		option.GroupTags("wallet"),
	)

	w.Get("", GetBalances(walletSvc)).With(
		option.OperationID("getBalances"),
		option.Summary("Get all wallet balances for current user"),
		option.Response(200, new(walletDTO.WalletListResponse)),
	)
	w.Get("/:asset_id", GetBalance(walletSvc)).With(
		option.OperationID("getBalance"),
		option.Summary("Get wallet balance for a specific asset"),
		option.Request(new(AssetIDParam)),
		option.Response(200, new(walletDTO.WalletResponse)),
	)
	w.Post("/deposit", Deposit(walletSvc)).With(
		option.OperationID("deposit"),
		option.Summary("Deposit funds to wallet"),
		option.Request(new(walletDTO.DepositDTO)),
		option.Response(200, new(walletDTO.WalletResponse)),
	)
	w.Post("/withdraw", Withdraw(walletSvc)).With(
		option.OperationID("withdraw"),
		option.Summary("Withdraw funds from wallet"),
		option.Request(new(walletDTO.WithdrawDTO)),
		option.Response(200, new(walletDTO.WalletResponse)),
	)
}
