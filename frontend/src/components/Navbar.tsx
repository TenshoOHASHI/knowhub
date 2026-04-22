'use client';
import Link from 'next/link';
import { NAV_LINKS } from '@/lib/routes';
import { useSidebar } from '@/context/SidebarContext';
import { FiChevronLeft, FiChevronRight } from 'react-icons/fi';
import { usePathname } from 'next/navigation';

export default function Navbar() {
  const { isOpen, toggle } = useSidebar();
  const pathname = usePathname();

  const isWikiPage = pathname.startsWith('/wiki');

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
          <Link href='/' className='text-3xl font-bold dark:text-stone-100'>
            knowhub
          </Link>
        </div>

        {/* 既存のリンク */}
        <div className='flex gap-6'>
          {NAV_LINKS.map((link) => {
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
        </div>
      </div>
    </nav>
  );
}
