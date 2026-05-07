'use client';

import Link from 'next/link';
import Image from 'next/image';
import { NAV_LINKS, NAV_LINKS_WITH_AUTH, LOGOUT_ROUTE } from '@/lib/routes';
import { useSidebar } from '@/context/SidebarContext';
import { FiChevronLeft, FiChevronRight } from 'react-icons/fi';
import { usePathname } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';

export default function Navbar() {
  const { isOpen, toggle } = useSidebar();
  const pathname = usePathname();
  const { isLoggedIn, logout } = useAuth();

  const isWikiPage = pathname.startsWith('/wiki');

  // isLoggedIn が変わると React が自動で再レンダー → リンクが切り替わる
  // useEffect は不要
  const links = isLoggedIn ? NAV_LINKS_WITH_AUTH : NAV_LINKS;

  return (
    <nav className='border-b border-zinc-300 bg-zinc-100/95 shadow-sm backdrop-blur dark:border-stone-800 dark:bg-stone-950/50'>
      <div className='flex w-full flex-col gap-3 px-3 py-3 sm:flex-row sm:items-center sm:justify-between sm:px-6 sm:py-4 lg:px-8'>
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
        </div>

        {/* ナビリンク */}
        <div className='nav-scrollbar -mx-3 flex w-[calc(100%+1.5rem)] gap-1 overflow-x-auto px-3 pb-1 sm:mx-0 sm:w-auto sm:gap-3 sm:overflow-visible sm:px-0 sm:pb-0 lg:gap-6'>
          {links.map((link) => {
            const Icon = link.icon;
            const isActive = pathname === link.href;
            return (
              <Link
                key={link.href}
                href={link.href}
                className={`flex shrink-0 items-center gap-1 rounded-md px-2 py-1.5 text-sm transition-colors sm:text-base ${
                  isActive
                    ? 'bg-white text-zinc-950 shadow-sm dark:bg-stone-800 dark:text-white'
                    : 'text-zinc-700 hover:bg-white hover:text-black dark:text-stone-300 dark:hover:bg-stone-800 dark:hover:text-white'
                }`}
              >
                <Icon className='h-4 w-4 shrink-0' />
                <span className='whitespace-nowrap'>{link.label}</span>
              </Link>
            );
          })}

          {/* ログアウトボタン（ログイン時のみ） */}
          {isLoggedIn && (
            <button
              onClick={logout}
              className='flex shrink-0 items-center gap-1 rounded-md px-2 py-1.5 text-sm text-zinc-700 transition-colors hover:bg-white hover:text-black dark:text-stone-300 dark:hover:bg-stone-800 dark:hover:text-white sm:text-base'
            >
              <LOGOUT_ROUTE.icon className='h-4 w-4 shrink-0' />
              <span className='whitespace-nowrap'>{LOGOUT_ROUTE.label}</span>
            </button>
          )}
        </div>
      </div>
    </nav>
  );
}
