import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { authService, type IUser } from "@/core/api";
import {
  setAccessToken,
  clearAccessToken,
  setOnAuthFailure,
} from "@/core/api/service";

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

  // Register auth failure callback so the service layer can trigger logout
  // without being coupled to the router.
  useEffect(() => {
    setOnAuthFailure(() => {
      clearAccessToken();
      setUser(null);
    });
  }, []);

  // Restore session on page load via refresh cookie
  useEffect(() => {
    authService
      .refresh()
      .then((res) => {
        setAccessToken(res.accessToken);
        // Now fetch user profile with the fresh access token
        return authService.getMe();
      })
      .then((u) => setUser(u))
      .catch(() => {
        clearAccessToken();
      })
      .finally(() => setIsLoading(false));
  }, []);

  const login = async (email: string, password: string) => {
    const res = await authService.login(email, password);
    setAccessToken(res.accessToken);
    setUser(res.user);
  };

  const register = async (email: string, password: string) => {
    const res = await authService.register(email, password);
    setAccessToken(res.accessToken);
    setUser(res.user);
  };

  const logout = async () => {
    try {
      await authService.logout();
    } catch {
      // ignore logout errors
    } finally {
      clearAccessToken();
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
