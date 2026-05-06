'use client';

import {
  createContext,
  useContext,
  useEffect,
  useState,
  ReactNode,
} from 'react';
import { checkAuth } from '@/lib/api';

interface AuthContextType {
  isLoggedIn: boolean | null;
  login: () => void;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType>({
  isLoggedIn: null,
  login: () => {},
  logout: async () => {},
});

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isLoggedIn, setIsLoggedIn] = useState<boolean | null>(null);

  // 初回時にapiを呼び出す
  useEffect(() => {
    // rewrites で Cookie が Gateway に転送され、認証結果を取得
    checkAuth()
      .then((ok) => setIsLoggedIn(ok))
      .catch(() => setIsLoggedIn(false));
  }, []);

  // ログイン成功後に呼ぶ → isLoggedIn を即座に true にする
  // ２回目以降ログインする際に、状態を更新
  function login() {
    setIsLoggedIn(true);
  }

  // HttpOnly Cookie は JS から削除できない → Route Handler 経由で削除
  async function logout() {
    await fetch('/api/auth/logout', { method: 'POST' });
    setIsLoggedIn(false);
  }

  return (
    <AuthContext value={{ isLoggedIn, login, logout }}>{children}</AuthContext>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
