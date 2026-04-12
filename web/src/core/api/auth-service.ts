import { Service } from "./service";
import { AUTH_ENDPOINT } from "./endpoint";
import type { IAuthResponse, IRefreshResponse, IUser } from "./types";

const service = new Service(AUTH_ENDPOINT.LOGIN);

export const authService = {
  login(email: string, password: string): Promise<IAuthResponse> {
    return service.post<IAuthResponse>(
      { email, password },
      AUTH_ENDPOINT.LOGIN
    );
  },

  register(email: string, password: string): Promise<IAuthResponse> {
    return service.post<IAuthResponse>(
      { email, password },
      AUTH_ENDPOINT.REGISTER
    );
  },

  logout(): Promise<{ message: string }> {
    return service.post<{ message: string }>({}, AUTH_ENDPOINT.LOGOUT);
  },

  getMe(): Promise<IUser> {
    return service.get<IUser>(AUTH_ENDPOINT.ME);
  },

  refresh(): Promise<IRefreshResponse> {
    return service.post<IRefreshResponse>({}, AUTH_ENDPOINT.REFRESH);
  },
};
