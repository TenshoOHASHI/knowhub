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
    <nav className='border-b border-black dark:border-stone-600'>
      <div className='w-full px-8 py-4 flex items-center justify-between'>
        <div className='flex items-center gap-2'>
          {/* サイドバートグルボタン */}
          {isWikiPage && (
            <button
              onClick={toggle}
              className='p-1 hover:bg-gray-100 dark:hover:bg-stone-800 rounded'
            >
              {isOpen ? (
                <FiChevronLeft size={24} />
              ) : (
                <FiChevronRight size={24} />
              )}
            </button>
          )}

          {/* ホームリンク */}
          <Link href='/' className='flex items-center -ml-4'>
            <Image
              src='/dark.png'
              alt='TENTEN STUDIO'
              width={68}
              height={68}
              className='w-10 h-10 rounded-2xl object-cover dark:invert'
            />
            <span className='text-2xl font-bold dark:text-stone-100'>
              TenHub
            </span>
          </Link>
        </div>

        {/* ナビリンク */}
        <div className='flex gap-6'>
          {links.map((link) => {
            const Icon = link.icon;
            return (
              <Link
                key={link.href}
                href={link.href}
                className='flex items-center text-gray-600 dark:text-stone-400 hover:text-black dark:hover:text-stone-100'
              >
                <Icon className='mr-1' />
                {link.label}
              </Link>
            );
          })}

          {/* ログアウトボタン（ログイン時のみ） */}
          {isLoggedIn && (
            <button
              onClick={logout}
              className='flex items-center text-gray-600 dark:text-stone-400 hover:text-black dark:hover:text-stone-100'
            >
              <LOGOUT_ROUTE.icon className='mr-1' />
              {LOGOUT_ROUTE.label}
            </button>
          )}
        </div>
      </div>
    </nav>
  );
}
