'use client';

import { useEffect, useState } from 'react';
import { FiCommand } from 'react-icons/fi';

const SHORTCUTS = [
  { key: 'h', label: 'ホーム', path: '/' },
  { key: 'w', label: 'Wiki', path: '/wiki' },
  { key: 'p', label: 'Portfolio', path: '/portfolio' },
  // { key: 'c', label: 'Chat', path: '/cht' },
] as const;

export default function KeyboardShortcuts() {
  const [showHelp, setShowHelp] = useState(false);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement).tagName;
      if (tag === 'INPUT' || tag === 'TEXTAREA') return;

      if (e.key === '?') {
        setShowHelp((prev) => !prev);
        return;
      }

      switch (e.key) {
        case 'h':
          window.location.href = '/';
          break;
        case 'w':
          window.location.href = '/wiki';
          break;
        case 'p':
          window.location.href = '/portfolio';
          break;
          // case 'c':
          // window.location.href = '/chat';
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  const isVisible = showHelp;

  return (
    <div className='group fixed bottom-1 right-4'>
      <button className='border border-black dark:border-stone-600 rounded-full w-10 h-10 flex items-center justify-center hover:bg-gray-100 dark:hover:bg-stone-800'>
        <FiCommand />
      </button>

      <div
        className={`absolute bottom-14 right-0 border border-black dark:border-stone-600 rounded-lg p-4 bg-white dark:bg-stone-900 shadow-lg min-w-max transition-opacity duration-200 ${isVisible ? 'opacity-100 visible' : 'opacity-0 invisible group-hover:opacity-100 group-hover:visible'}`}
      >
        <h3 className='font-semibold mb-2 text-sm whitespace-nowrap'>
          ショートカット
        </h3>
        <ul className='space-y-1 text-sm whitespace-nowrap'>
          {SHORTCUTS.map((s) => (
            <li key={s.key} className='flex justify-between gap-4'>
              <span className='text-gray-600 dark:text-stone-400'>
                {s.label}
              </span>
              <kbd className='border border-black dark:border-stone-600 px-2 py-0.5 rounded text-xs'>
                {s.key}
              </kbd>
            </li>
          ))}
          <li className='flex justify-between gap-4'>
            <span className='text-gray-600 dark:text-stone-400'>
              このメニュー
            </span>
            <kbd className='border border-black dark:border-stone-600 px-2 py-0.5 rounded text-xs'>
              ?
            </kbd>
          </li>
        </ul>
      </div>
    </div>
  );
}
