import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import type { IUser } from "@/core/api";

interface AuthContextValue {
  user: IUser | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<IUser | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem("access_token");
    if (!token) {
      setIsLoading(false);
      return;
    }
    // Lazy import to avoid circular deps
    import("@/features/auth/data").then(({ AuthModel }) => {
      AuthModel.getMe()
        .then((u) => setUser(u))
        .catch(() => {
          localStorage.removeItem("access_token");
          localStorage.removeItem("refresh_token");
        })
        .finally(() => setIsLoading(false));
    });
  }, []);

  const login = async (email: string, password: string) => {
    const { AuthModel } = await import("@/features/auth/data");
    const res = await AuthModel.login(email, password);
    localStorage.setItem("access_token", res.access_token);
    localStorage.setItem("refresh_token", res.refresh_token);
    setUser(res.user);
  };

  const register = async (email: string, password: string) => {
    const { AuthModel } = await import("@/features/auth/data");
    const res = await AuthModel.register(email, password);
    localStorage.setItem("access_token", res.access_token);
    localStorage.setItem("refresh_token", res.refresh_token);
    setUser(res.user);
  };

  const logout = async () => {
    try {
      const { AuthModel } = await import("@/features/auth/data");
      await AuthModel.logout();
    } catch {
      // ignore logout errors
    } finally {
      localStorage.removeItem("access_token");
      localStorage.removeItem("refresh_token");
      setUser(null);
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        isAuthenticated: !!user,
        login,
        register,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}
