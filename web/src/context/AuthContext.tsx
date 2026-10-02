import React, { createContext, useContext, useState, useEffect } from 'react';
import { User, Role, TokenPair, LoginPayload, RegisterPayload } from '../types/auth';

interface AuthContextType {
  user: User | null;
  tokens: TokenPair | null;
  permissions: string[];
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (payload: LoginPayload) => Promise<Role>;
  register: (payload: RegisterPayload) => Promise<Role>;
  logout: () => Promise<void>;
  hasPermission: (perm: string) => boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [tokens, setTokens] = useState<TokenPair | null>(null);
  const [permissions, setPermissions] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  useEffect(() => {
    const savedTokens = localStorage.getItem('logiflows_tokens');
    const savedUser = localStorage.getItem('logiflows_user');
    const savedPerms = localStorage.getItem('logiflows_permissions');

    if (savedTokens && savedUser) {
      try {
        setTokens(JSON.parse(savedTokens));
        setUser(JSON.parse(savedUser));
        if (savedPerms) {
          setPermissions(JSON.parse(savedPerms));
        }
      } catch (e) {
        console.error('Failed to parse cached auth state', e);
        localStorage.clear();
      }
    }
    setIsLoading(false);
  }, []);

  const login = async (payload: LoginPayload): Promise<Role> => {
    setIsLoading(true);
    try {
      const res = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      const json = await res.json();
      if (!res.ok) {
        throw new Error(json.error?.message || 'Login failed. Please check your credentials.');
      }

      const { user: authedUser, tokens: authTokens, permissions: authPerms } = json.data;

      setUser(authedUser);
      setTokens(authTokens);
      setPermissions(authPerms || []);

      localStorage.setItem('logiflows_user', JSON.stringify(authedUser));
      localStorage.setItem('logiflows_tokens', JSON.stringify(authTokens));
      localStorage.setItem('logiflows_permissions', JSON.stringify(authPerms || []));

      return authedUser.role;
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (payload: RegisterPayload): Promise<Role> => {
    setIsLoading(true);
    try {
      const res = await fetch('/api/v1/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      const json = await res.json();
      if (!res.ok) {
        throw new Error(json.error?.message || 'Registration failed.');
      }

      const { user: authedUser, tokens: authTokens, permissions: authPerms } = json.data;

      setUser(authedUser);
      setTokens(authTokens);
      setPermissions(authPerms || []);

      localStorage.setItem('logiflows_user', JSON.stringify(authedUser));
      localStorage.setItem('logiflows_tokens', JSON.stringify(authTokens));
      localStorage.setItem('logiflows_permissions', JSON.stringify(authPerms || []));

      return authedUser.role;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    if (tokens?.refresh_token) {
      try {
        await fetch('/api/v1/auth/logout', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${tokens.access_token}`,
          },
          body: JSON.stringify({ refresh_token: tokens.refresh_token }),
        });
      } catch (err) {
        console.error('Logout error on backend', err);
      }
    }

    setUser(null);
    setTokens(null);
    setPermissions([]);
    localStorage.removeItem('logiflows_user');
    localStorage.removeItem('logiflows_tokens');
    localStorage.removeItem('logiflows_permissions');
  };

  const hasPermission = (perm: string): boolean => {
    return permissions.includes(perm);
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        tokens,
        permissions,
        isAuthenticated: !!user,
        isLoading,
        login,
        register,
        logout,
        hasPermission,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
