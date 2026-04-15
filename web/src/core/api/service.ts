import axios, { type AxiosInstance, type InternalAxiosRequestConfig } from "axios";
import type { IHttpError } from "./types";
import { AUTH_ENDPOINT } from "./endpoint";

// In-memory access token (never in localStorage)
let accessToken: string | null = null;

export function getAccessToken(): string | null {
  return accessToken;
}

export function setAccessToken(token: string | null) {
  accessToken = token;
}

export function clearAccessToken() {
  accessToken = null;
}

// Callback for auth failure — set by AuthProvider to handle logout without coupling to router
let onAuthFailure: (() => void) | null = null;

export function setOnAuthFailure(cb: () => void) {
  onAuthFailure = cb;
}

// Silent refresh queue
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (token: string) => void;
  reject: (err: unknown) => void;
}> = [];

function processQueue(error: unknown, token: string | null) {
  failedQueue.forEach(({ resolve, reject }) =>
    error ? reject(error) : resolve(token!)
  );
  failedQueue = [];
}

export class Service {
  private endpoint: string;
  private client: AxiosInstance;

  constructor(endpoint: string) {
    this.endpoint = endpoint;
    this.client = axios.create({
      baseURL: import.meta.env.VITE_API_BASE_URL || "",
      headers: { "Content-Type": "application/json" },
      withCredentials: true,
    });

    // Attach access token from memory
    this.client.interceptors.request.use((config) => {
      const token = getAccessToken();
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    });

    // Silent refresh on 401
    this.client.interceptors.response.use(
      (res) => res,
      async (error) => {
        const original = error.config as InternalAxiosRequestConfig & {
          _retry?: boolean;
        };

        // Don't retry refresh endpoint itself, non-401 errors, or already retried
        if (
          error.response?.status !== 401 ||
          original._retry ||
          original.url === AUTH_ENDPOINT.REFRESH
        ) {
          const httpError: IHttpError = {
            httpCode: error.response?.status ?? 500,
            message:
              error.response?.data?.error?.message ?? "Something went wrong",
          };
          return Promise.reject(httpError);
        }

        // Queue concurrent requests while refreshing
        if (isRefreshing) {
          return new Promise<string>((resolve, reject) => {
            failedQueue.push({ resolve, reject });
          }).then((token) => {
            original.headers.Authorization = `Bearer ${token}`;
            return this.client(original);
          });
        }

        original._retry = true;
        isRefreshing = true;

        try {
          const { data } = await this.client.post<{
            data: { accessToken: string };
          }>(AUTH_ENDPOINT.REFRESH);

          const newToken = data.data.accessToken;
          setAccessToken(newToken);
          processQueue(null, newToken);

          original.headers.Authorization = `Bearer ${newToken}`;
          return this.client(original);
        } catch (refreshError) {
          processQueue(refreshError, null);
          clearAccessToken();
          onAuthFailure?.();
          return Promise.reject(refreshError);
        } finally {
          isRefreshing = false;
        }
      }
    );
  }

  private resolve<T>(response: { data: { data: T } }): T {
    return response.data?.data;
  }

  async get<T>(url?: string, params?: object): Promise<T> {
    const res = await this.client.get(url || this.endpoint, { params });
    return this.resolve<T>(res);
  }

  async post<T>(payload: object, url?: string): Promise<T> {
    const res = await this.client.post(url || this.endpoint, payload);
    return this.resolve<T>(res);
  }

  async delete<T>(url: string): Promise<T> {
    const res = await this.client.delete(url);
    return this.resolve<T>(res);
  }
}
