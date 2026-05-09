'use client';

import { useTheme } from '@/context/ThemeContext';

export default function ThemeToggle() {
  const { theme, toggleTheme } = useTheme();

  return (
    <button
      onClick={toggleTheme}
      className='shrink-0 rounded-md p-2 text-zinc-700 hover:bg-white hover:text-black dark:text-stone-300 dark:hover:bg-stone-800 dark:hover:text-white transition-colors'
      title={theme === 'light' ? 'ダークモードに切替' : 'ライトモードに切替'}
    >
      {theme === 'light' ? (
        <svg className='h-5 w-5' fill='none' stroke='currentColor' viewBox='0 0 24 24'>
          <path strokeLinecap='round' strokeLinejoin='round' strokeWidth={2} d='M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z' />
        </svg>
      ) : (
        <svg className='h-5 w-5' fill='none' stroke='currentColor' viewBox='0 0 24 24'>
          <path strokeLinecap='round' strokeLinejoin='round' strokeWidth={2} d='M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z' />
        </svg>
      )}
    </button>
  );
}
