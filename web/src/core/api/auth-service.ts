import {
  postApiV1AuthLogin,
  postApiV1AuthRegister,
  postApiV1AuthLogout,
  getApiV1AuthMe,
  postApiV1AuthRefresh,
} from "./generated/auth/auth";
import type { IAuthResponse, IRefreshResponse, IUser } from "./types";

export const authService = {
  login(email: string, password: string): Promise<IAuthResponse> {
    return postApiV1AuthLogin({ email, password }) as Promise<IAuthResponse>;
  },

  register(email: string, password: string): Promise<IAuthResponse> {
    return postApiV1AuthRegister({ email, password }) as Promise<IAuthResponse>;
  },

  logout(): Promise<{ message: string }> {
    return postApiV1AuthLogout() as Promise<{ message: string }>;
  },

  getMe(): Promise<IUser> {
    return getApiV1AuthMe() as Promise<IUser>;
  },

  refresh(): Promise<IRefreshResponse> {
    return postApiV1AuthRefresh() as Promise<IRefreshResponse>;
  },
};
