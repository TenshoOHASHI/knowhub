'use client';

import { createContext, useContext, useEffect, useRef, useState } from 'react';

type Theme = 'light' | 'dark';

interface ThemeContextType {
  theme: Theme;
  toggleTheme: () => void;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

const THEME_KEY = 'tenhub-theme';

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = useState<Theme>('light');
  const isInitialized = useRef(false);

  useEffect(() => {
    if (isInitialized.current) return;
    isInitialized.current = true;

    // マウント時にlocalStorageからテーマを読み込む
    const saved = localStorage.getItem(THEME_KEY) as Theme;
    const initialTheme: Theme = saved === 'light' || saved === 'dark' ? saved : 'light';

    // HTML要素にdarkクラスを設定
    const root = document.documentElement;
    root.classList.toggle('dark', initialTheme === 'dark');

    // Stateを更新してUIを反映
    setThemeState(initialTheme);
  }, []);

  useEffect(() => {
    // 初回マウント後のみテーマ変更を監視
    if (!isInitialized.current) return;

    const root = document.documentElement;
    const isDark = theme === 'dark';
    root.classList.toggle('dark', isDark);
    localStorage.setItem(THEME_KEY, theme);
  }, [theme]);

  const toggleTheme = () => {
    setThemeState((prev) => (prev === 'light' ? 'dark' : 'light'));
  };

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within ThemeProvider');
  }
  return context;
}
