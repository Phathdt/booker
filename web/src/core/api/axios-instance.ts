import Axios, { type AxiosRequestConfig } from "axios";
import { getAccessToken, setAccessToken, clearAccessToken } from "./service";
import { AUTH_ENDPOINT } from "./endpoint";

// Shared axios instance with auth interceptors (used by orval-generated hooks)
export const AXIOS_INSTANCE = Axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "",
  headers: { "Content-Type": "application/json" },
  withCredentials: true,
});

// Attach access token
AXIOS_INSTANCE.interceptors.request.use((config) => {
  const token = getAccessToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Silent refresh on 401
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (token: string) => void;
  reject: (err: unknown) => void;
}> = [];

function processQueue(error: unknown, token: string | null) {
  failedQueue.forEach(({ resolve, reject }) =>
    error ? reject(error) : resolve(token!),
  );
  failedQueue = [];
}

AXIOS_INSTANCE.interceptors.response.use(
  (res) => res,
  async (error) => {
    const original = error.config as AxiosRequestConfig & { _retry?: boolean };

    if (
      error.response?.status !== 401 ||
      original._retry ||
      original.url === AUTH_ENDPOINT.REFRESH
    ) {
      return Promise.reject(error);
    }

    if (isRefreshing) {
      return new Promise<string>((resolve, reject) => {
        failedQueue.push({ resolve, reject });
      }).then((token) => {
        original.headers = { ...original.headers, Authorization: `Bearer ${token}` };
        return AXIOS_INSTANCE(original);
      });
    }

    original._retry = true;
    isRefreshing = true;

    try {
      const { data } = await AXIOS_INSTANCE.post<{ data: { access_token: string } }>(
        AUTH_ENDPOINT.REFRESH,
      );
      const newToken = data.data.access_token;
      setAccessToken(newToken);
      processQueue(null, newToken);
      original.headers = { ...original.headers, Authorization: `Bearer ${newToken}` };
      return AXIOS_INSTANCE(original);
    } catch (refreshError) {
      processQueue(refreshError, null);
      clearAccessToken();
      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  },
);

// Orval custom instance — must match mutator signature
export const axiosInstance = <T>(config: AxiosRequestConfig): Promise<T> => {
  return AXIOS_INSTANCE(config).then(({ data }) => data);
};

export default axiosInstance;
