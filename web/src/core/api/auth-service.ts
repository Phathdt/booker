import {
  login as apiLogin,
  register as apiRegister,
  logout as apiLogout,
  getMe as apiGetMe,
  refreshToken as apiRefreshToken,
} from "./generated/auth/auth";
import type { IAuthResponse, IRefreshResponse, IUser } from "./types";

export const authService = {
  login(email: string, password: string): Promise<IAuthResponse> {
    return apiLogin({ email, password }) as Promise<IAuthResponse>;
  },

  register(email: string, password: string): Promise<IAuthResponse> {
    return apiRegister({ email, password }) as Promise<IAuthResponse>;
  },

  logout(): Promise<{ message: string }> {
    return apiLogout() as Promise<{ message: string }>;
  },

  getMe(): Promise<IUser> {
    return apiGetMe() as Promise<IUser>;
  },

  refresh(): Promise<IRefreshResponse> {
    return apiRefreshToken() as Promise<IRefreshResponse>;
  },
};
