import { useMutation } from "@tanstack/react-query";
import { AuthModel } from "../models";

export function useMutationLogin() {
  return useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) =>
      AuthModel.login(email, password),
  });
}

export function useMutationRegister() {
  return useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) =>
      AuthModel.register(email, password),
  });
}
