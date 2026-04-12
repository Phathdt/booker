import { useQuery } from "@tanstack/react-query";
import { AuthModel } from "../models";

export const QUERY_KEYS = {
  AUTH: { ME: "auth-me" },
};

export function useQueryMe() {
  return useQuery({
    queryKey: [QUERY_KEYS.AUTH.ME],
    queryFn: () => AuthModel.getMe(),
    enabled: !!localStorage.getItem("access_token"),
  });
}
