'use client';
import { useState } from 'react';
import Link from 'next/link';
import { FiSearch } from 'react-icons/fi';
import type { Article } from '@/lib/types';

export default function WikiClient({ articles }: { articles: Article[] }) {
  const [query, setQuery] = useState('');
  const filtered = articles.filter((a) => {
    const q = query.toLowerCase();
    return (
      a.title.toLocaleLowerCase().includes(q) ||
      a.content.toLocaleLowerCase().includes(q)
    );
  });

  return (
    <div className='max-w-4xl mx-auto p-6'>
      {/* 検索バー（アイコン＋入力欄が同じ relative の中） */}
      <div className='relative mb-6'>
        <FiSearch className='absolute left-3 top-1/2 -translate-y-1/2 text-gray-400' />
        <input
          type='text'
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder='記事を検索...'
          className='w-full border border-black dark:border-stone-600 rounded-lg pl-10 pr-4 py-2 bg-transparent'
        />
      </div>

      <h1 className='text-3xl font-bold mb-6 '>
        <span className='inline-block border-b-2 border-red-500/40 pb-1'>
          Knowledge
        </span>
      </h1>
      {filtered.length == 0 && <p>記事が存在しません。</p>}
      <div className='space-y-4'>
        {filtered.map((a) => (
          <Link href={`/wiki/${a.id}`} key={a.id} className='block'>
            <div className='border border-black dark:border-stone-600 rounded-lg p-4 hover:shadow-md'>
              <h2 className='text-xl font-semibold'>{a.title}</h2>
              <p className='text-gray-600 dark:text-stone-400 mt-2 line-clamp-2'>
                {a.content}
              </p>
              <p className='text-sm text-gray-400 mt-2'>
                {a.created_at
                  ? new Date(a.created_at.seconds * 1000).toLocaleDateString(
                      'ja-JP',
                    )
                  : ''}
              </p>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
