import { Service } from "@/core/api/service";
import { AUTH_ENDPOINT } from "@/core/api/endpoint";
import type { IAuthResponse, IUser } from "@/core/api/types";

export class AuthModel {
  private static service = new Service(AUTH_ENDPOINT.LOGIN);

  static login(email: string, password: string): Promise<IAuthResponse> {
    return AuthModel.service.post<IAuthResponse>(
      { email, password },
      AUTH_ENDPOINT.LOGIN
    );
  }

  static register(email: string, password: string): Promise<IAuthResponse> {
    return AuthModel.service.post<IAuthResponse>(
      { email, password },
      AUTH_ENDPOINT.REGISTER
    );
  }

  static logout(): Promise<{ message: string }> {
    return AuthModel.service.post<{ message: string }>(
      {},
      AUTH_ENDPOINT.LOGOUT
    );
  }

  static getMe(): Promise<IUser> {
    return AuthModel.service.get<IUser>(AUTH_ENDPOINT.ME);
  }
}
