'use client';

import { useContext, createContext, useState, ReactNode } from 'react';

interface SidebarContextType {
  isOpen: boolean;
  toggle: () => void;
}

// Contextを作成
const SidebarContext = createContext<SidebarContextType | undefined>(undefined);

// 親要素から子要素にコンテンツを共有
export function SidebarProvider({ children }: { children: ReactNode }) {
  const [isOpen, setIsOpen] = useState(true); // デフォルトは開いている
  const toggle = () => setIsOpen((prev) => !prev);

  return <SidebarContext value={{ isOpen, toggle }}>{children}</SidebarContext>;
}

// カスタムフック
export function useSidebar() {
  // コンテキストを取り出す
  const context = useContext(SidebarContext);
  if (!context) {
    throw new Error('useSidebar must be used within SidebarProvider');
  }
  return context;
}
