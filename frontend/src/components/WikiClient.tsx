'use client';
import { useState } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { FiSearch, FiCalendar, FiFileText, FiLock } from 'react-icons/fi';
import type { Article } from '@/lib/types';
import { motion } from 'motion/react';

// Markdown記号を除去してプレーンテキストにする
function stripMarkdown(md: string): string {
  return md
    .replace(/```[\s\S]*?```/g, '') // コードブロック除去
    .replace(/`([^`]+)`/g, '$1') // インラインコード
    .replace(/!\[.*?\]\(.*?\)/g, '') // 画像
    .replace(/\[([^\]]+)\]\(.*?\)/g, '$1') // リンク → テキストだけ
    .replace(/#{1,6}\s/g, '') // 見出し記号
    .replace(/(\*{1,3}|_{1,3})(.*?)\1/g, '$2') // 太字・斜体
    .replace(/>\s.*/g, '') // 引用
    .replace(/^[-*+]\s/gm, '') // リスト記号
    .replace(/^\|.+\|$/gm, '') // テーブル行（| ... |）
    .replace(/^---+$/gm, '') // 水平線
    .replace(/\n{2,}/g, ' ') // 改行をスペースに
    .trim();
}

export default function WikiClient({ articles }: { articles: Article[] }) {
  const [query, setQuery] = useState('');
  const searchParams = useSearchParams();
  const categoryId = searchParams.get('category');

  const filtered = articles.filter((a) => {
    const q = query.toLowerCase();
    const matchesQuery =
      a.title.toLowerCase().includes(q) || a.content.toLowerCase().includes(q);
    const matchesCategory = !categoryId || a.category_id === categoryId;
    return matchesQuery && matchesCategory;
  });

  return (
    <div className='max-w-4xl mx-auto p-6'>
      {/* 検索バー */}
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

      <h1 className='text-4xl font-bold mb-6'>
        <span className='inline-block border-b-2 border-red-500/40 pb-1'>
          <motion.span
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
          >
            <span className='bg-gradient-to-r from-stone-600 to-stone-300 bg-clip-text text-transparent'>
              Knowledge
            </span>
          </motion.span>
        </span>
      </h1>

      {filtered.length == 0 && (
        <div className='text-center py-12'>
          <FiFileText
            className='mx-auto mb-3 text-gray-300 dark:text-stone-600'
            size={48}
          />
          <p className='text-gray-400'>記事が存在しません</p>
        </div>
      )}

      <div className='space-y-3'>
        {filtered.map((a) => (
          <Link href={`/wiki/${a.id}`} key={a.id} className='block group'>
            <div
              className='border border-black/10 dark:border-stone-700 rounded-lg p-5
              hover:border-black/20 dark:hover:border-stone-500
              hover:shadow-md transition-all duration-200 overflow-hidden'
            >
              <h2 className='text-lg font-semibold group-hover:text-red-500/70 transition-colors flex items-center gap-2'>
                {a.visibility === 'locked' && (
                  <span className='inline-flex items-center justify-center w-5 h-5 rounded bg-stone-200 dark:bg-stone-700 shrink-0'>
                    <FiLock
                      size={11}
                      className='text-stone-500 dark:text-stone-400'
                    />
                  </span>
                )}
                <span className='min-w-0 break-all'>{a.title}</span>
              </h2>
              <p className='text-gray-500 dark:text-stone-400 mt-2 line-clamp-2 text-sm leading-relaxed'>
                {a.visibility === 'locked'
                  ? 'この記事は限定公開です'
                  : stripMarkdown(a.content).slice(0, 150)}
              </p>
              <div className='flex items-center gap-1.5 mt-3 text-xs text-gray-400 dark:text-stone-500'>
                <FiCalendar size={12} />
                <span>
                  {a.created_at
                    ? new Date(a.created_at.seconds * 1000).toLocaleDateString(
                        'ja-JP',
                      )
                    : ''}
                </span>
              </div>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
