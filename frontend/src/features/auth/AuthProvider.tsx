import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { queryClient } from "@/app/queryClient";
import { getMe, login } from "./api";
import {
  clearAuthSession,
  getAuthToken,
  getStoredUser,
  saveAuthTokens,
  saveAuthUser,
} from "./session";
import type { AuthUser, LoginInput } from "./types";

type AuthContextValue = {
  user: AuthUser | null;
  token: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  loginWithEmail: (input: LoginInput) => Promise<AuthUser>;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => getAuthToken());
  const [user, setUser] = useState<AuthUser | null>(() => getStoredUser());
  const [isLoading, setIsLoading] = useState(Boolean(getAuthToken()));

  const logout = useCallback(() => {
    clearAuthSession();
    setToken(null);
    setUser(null);
  }, []);

  useEffect(() => {
    if (!token) {
      setIsLoading(false);
      return;
    }

    let active = true;
    setIsLoading(true);

    getMe()
      .then((me) => {
        if (!active) return;
        saveAuthUser(me);
        setUser(me);
      })
      .catch(() => {
        if (!active) return;
        logout();
      })
      .finally(() => {
        if (active) setIsLoading(false);
      });

    return () => {
      active = false;
    };
  }, [logout, token]);

  const loginWithEmail = useCallback(async (input: LoginInput) => {
    const res = await login(input);
    // Si en esta misma pestaña había datos en caché de una sesión anterior
    // (otro usuario u otro rol, por ejemplo al probar varios roles seguidos
    // sin recargar la página), hay que descartarlos antes de mostrar la
    // nueva sesión: la queryKey de varias pantallas (ej. Panel General) no
    // incluye el usuario, así que un valor viejo podría mostrarse un
    // instante antes de que la primera consulta de la sesión nueva termine.
    queryClient.clear();
    saveAuthTokens(res.id_token, res.refresh_token);
    setToken(res.id_token);

    const me = await getMe();
    saveAuthUser(me);
    setUser(me);
    return me;
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      token,
      isLoading,
      isAuthenticated: Boolean(token && user),
      loginWithEmail,
      logout,
    }),
    [isLoading, loginWithEmail, logout, token, user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used inside AuthProvider");
  }
  return ctx;
}
