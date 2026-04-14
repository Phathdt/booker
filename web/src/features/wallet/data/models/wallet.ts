import {
  getApiV1Wallet,
  getApiV1WalletAssetId,
  postApiV1WalletDeposit,
  postApiV1WalletWithdraw,
} from "@/core/api/generated/wallet/wallet";
import type { IWallet } from "@/core/api/types";

interface WalletListResponse {
  wallets: IWallet[];
}

export class WalletModel {
  static getAll(): Promise<WalletListResponse> {
    return getApiV1Wallet() as Promise<WalletListResponse>;
  }

  static getByAsset(assetId: string): Promise<IWallet> {
    return getApiV1WalletAssetId(assetId) as Promise<IWallet>;
  }

  static deposit(assetId: string, amount: string): Promise<IWallet> {
    return postApiV1WalletDeposit({ asset_id: assetId, amount }) as Promise<IWallet>;
  }

  static withdraw(assetId: string, amount: string): Promise<IWallet> {
    return postApiV1WalletWithdraw({ asset_id: assetId, amount }) as Promise<IWallet>;
  }
}
