import axios, { type AxiosInstance } from "axios";
import type { IHttpError } from "./types";

export class Service {
  private endpoint: string;

  constructor(endpoint: string) {
    this.endpoint = endpoint;
  }

  private createClient(): AxiosInstance {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
    const token = localStorage.getItem("access_token");
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }
    const instance = axios.create({
      baseURL: import.meta.env.VITE_API_BASE_URL || "",
      headers,
    });
    instance.interceptors.response.use(
      (res) => res,
      (error) => {
        const httpError: IHttpError = {
          httpCode: error.response?.status ?? 500,
          message:
            error.response?.data?.error?.message ?? "Something went wrong",
        };
        if (error.response?.status === 401) {
          localStorage.removeItem("access_token");
          localStorage.removeItem("refresh_token");
          window.location.href = "/login";
        }
        return Promise.reject(httpError);
      }
    );
    return instance;
  }

  private resolve<T>(response: { data: { data: T } }): T {
    return response.data?.data;
  }

  async get<T>(url?: string, params?: object): Promise<T> {
    const res = await this.createClient().get(url || this.endpoint, { params });
    return this.resolve<T>(res);
  }

  async post<T>(payload: object, url?: string): Promise<T> {
    const res = await this.createClient().post(url || this.endpoint, payload);
    return this.resolve<T>(res);
  }

  async delete<T>(url: string): Promise<T> {
    const res = await this.createClient().delete(url);
    return this.resolve<T>(res);
  }
}
