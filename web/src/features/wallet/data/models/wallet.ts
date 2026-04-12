import { Service } from "@/core/api/service";
import { WALLET_ENDPOINT } from "@/core/api/endpoint";
import type { IWallet } from "@/core/api/types";

interface WalletListResponse {
  wallets: IWallet[];
}

export class WalletModel {
  private static service = new Service(WALLET_ENDPOINT.LIST);

  static getAll(): Promise<WalletListResponse> {
    return WalletModel.service.get<WalletListResponse>(WALLET_ENDPOINT.LIST);
  }

  static getByAsset(assetId: string): Promise<IWallet> {
    return WalletModel.service.get<IWallet>(`${WALLET_ENDPOINT.LIST}/${assetId}`);
  }

  static deposit(assetId: string, amount: string): Promise<IWallet> {
    return WalletModel.service.post<IWallet>(
      { asset_id: assetId, amount },
      WALLET_ENDPOINT.DEPOSIT
    );
  }

  static withdraw(assetId: string, amount: string): Promise<IWallet> {
    return WalletModel.service.post<IWallet>(
      { asset_id: assetId, amount },
      WALLET_ENDPOINT.WITHDRAW
    );
  }
}
