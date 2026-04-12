import { authService } from "@/core/api";

// Re-export core authService as AuthModel for feature-level usage
export const AuthModel = authService;
