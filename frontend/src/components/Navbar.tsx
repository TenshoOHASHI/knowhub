'use client';

import Link from 'next/link';
import Image from 'next/image';
import { NAV_LINKS, NAV_LINKS_WITH_AUTH, LOGOUT_ROUTE } from '@/lib/routes';
import { useSidebar } from '@/context/SidebarContext';
import { FiChevronLeft, FiChevronRight, FiMenu, FiX } from 'react-icons/fi';
import { usePathname } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import ThemeToggle from '@/components/ThemeToggle';
import { useState } from 'react';

export default function Navbar() {
  const { isOpen, toggle } = useSidebar();
  const pathname = usePathname();
  const { isLoggedIn, logout } = useAuth();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  const isWikiPage = pathname.startsWith('/wiki');

  const links = isLoggedIn ? NAV_LINKS_WITH_AUTH : NAV_LINKS;

  return (
    <nav className='border-b border-zinc-300 bg-zinc-100/95 shadow-sm backdrop-blur dark:border-stone-800 dark:bg-stone-950/50'>
      <div className='flex w-full flex-col gap-3 px-3 py-3 sm:flex-row sm:items-center sm:justify-between sm:px-6 sm:py-4 lg:px-8'>
        {/* ヘッダー部分：ロゴとハンバーガーボタン */}
        <div className='flex min-w-0 items-center justify-between gap-2 sm:justify-start'>
          <div className='flex min-w-0 items-center gap-2'>
            {/* サイドバートグルボタン */}
            {isWikiPage && (
              <button
                onClick={toggle}
                aria-label={isOpen ? 'サイドバーを閉じる' : 'サイドバーを開く'}
                className='shrink-0 rounded p-1 text-zinc-700 hover:bg-white hover:text-black dark:text-stone-200 dark:hover:bg-stone-800 dark:hover:text-white'
              >
                {isOpen ? (
                  <FiChevronLeft className='h-6 w-6' />
                ) : (
                  <FiChevronRight className='h-6 w-6' />
                )}
              </button>
            )}

            {/* ホームリンク */}
            <Link href='/' className='flex min-w-0 items-center gap-2'>
              <Image
                src='/dark.png'
                alt='TENTEN STUDIO'
                width={68}
                height={68}
                className='h-9 w-9 shrink-0 rounded-2xl object-cover dark:invert sm:h-10 sm:w-10'
              />
              <span className='truncate text-xl font-bold text-zinc-950 dark:text-stone-100 sm:text-2xl'>
                TenHub
              </span>
            </Link>
          </div>

          {/* モバイル用：ハンバーガーボタンとテーマ切り替え */}
          <div className='flex items-center gap-2 sm:hidden'>
            <ThemeToggle />
            <button
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
              aria-label={isMobileMenuOpen ? 'メニューを閉じる' : 'メニューを開く'}
              className='shrink-0 rounded p-2 text-zinc-700 hover:bg-white hover:text-black dark:text-stone-300 dark:hover:bg-stone-800 dark:hover:text-white'
            >
              {isMobileMenuOpen ? (
                <FiX className='h-6 w-6' />
              ) : (
                <FiMenu className='h-6 w-6' />
              )}
            </button>
          </div>
        </div>

        {/* デスクトップ用：ナビリンク（横並び） */}
        <div className='hidden gap-3 lg:gap-6 sm:flex'>
          {links.map((link) => {
            const Icon = link.icon;
            const isActive = pathname === link.href;
            return (
              <Link
                key={link.href}
                href={link.href}
                className={`flex items-center gap-1 rounded-md px-3 py-1.5 text-base transition-colors ${
                  isActive
                    ? 'bg-white text-zinc-950 shadow-sm dark:bg-stone-800 dark:text-white'
                    : 'text-zinc-700 hover:bg-white hover:text-black dark:text-stone-300 dark:hover:bg-stone-800 dark:hover:text-white'
                }`}
                onClick={() => setIsMobileMenuOpen(false)}
              >
                <Icon className='h-4 w-4 shrink-0' />
                <span>{link.label}</span>
              </Link>
            );
          })}

          {/* ログアウトボタン（ログイン時のみ） */}
          {isLoggedIn && (
            <button
              onClick={logout}
              className='flex items-center gap-1 rounded-md px-3 py-1.5 text-base text-zinc-700 transition-colors hover:bg-white hover:text-black dark:text-stone-300 dark:hover:bg-stone-800 dark:hover:text-white'
            >
              <LOGOUT_ROUTE.icon className='h-4 w-4 shrink-0' />
              <span>{LOGOUT_ROUTE.label}</span>
            </button>
          )}

          {/* デスクトップ用：テーマ切り替え */}
          <ThemeToggle />
        </div>
      </div>

      {/* モバイル用：ドロップダウンメニュー */}
      {isMobileMenuOpen && (
        <div className='border-t border-zinc-300 bg-zinc-100/95 backdrop-blur dark:border-stone-800 dark:bg-stone-950/50 sm:hidden'>
          <div className='flex flex-col gap-1 px-3 py-3'>
            {links.map((link) => {
              const Icon = link.icon;
              const isActive = pathname === link.href;
              return (
                <Link
                  key={link.href}
                  href={link.href}
                  className={`flex items-center gap-3 rounded-md px-3 py-2.5 text-sm transition-colors ${
                    isActive
                      ? 'bg-white text-zinc-950 shadow-sm dark:bg-stone-800 dark:text-white'
                      : 'text-zinc-700 hover:bg-white hover:text-black dark:text-stone-300 dark:hover:bg-stone-800 dark:hover:text-white'
                  }`}
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  <Icon className='h-5 w-5 shrink-0' />
                  <span>{link.label}</span>
                </Link>
              );
            })}

            {/* ログアウトボタン（ログイン時のみ） */}
            {isLoggedIn && (
              <button
                onClick={() => {
                  logout();
                  setIsMobileMenuOpen(false);
                }}
                className='flex items-center gap-3 rounded-md px-3 py-2.5 text-sm text-zinc-700 transition-colors hover:bg-white hover:text-black dark:text-stone-300 dark:hover:bg-stone-800 dark:hover:text-white'
              >
                <LOGOUT_ROUTE.icon className='h-5 w-5 shrink-0' />
                <span>{LOGOUT_ROUTE.label}</span>
              </button>
            )}
          </div>
        </div>
      )}
    </nav>
  );
}
