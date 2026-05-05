'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { FiBookmark, FiArrowLeft } from 'react-icons/fi';
import { getFingerprint } from '@/lib/fingerprint';
import { listSavedArticles } from '@/lib/api';
import type { Article } from '@/lib/types';

export default function SavedArticlesPage() {
  const [articles, setArticles] = useState<Article[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const fp = await getFingerprint();
        const data = await listSavedArticles(fp);
        setArticles(data.articles || []);
      } catch {
        // Silently fail
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  return (
    <div className='max-w-4xl mx-auto px-6 py-8'>
      <div className='flex items-center gap-3 mb-8'>
        <Link
          href='/wiki'
          className='inline-flex items-center text-sm text-stone-500 hover:text-black dark:hover:text-stone-100'
        >
          <FiArrowLeft className='mr-1' size={14} />
          Wiki
        </Link>
      </div>

      <div className='flex items-center gap-2 mb-6'>
        <FiBookmark size={20} className='text-blue-500' />
        <h1 className='text-2xl font-bold'>保存した記事</h1>
      </div>

      {loading ? (
        <div className='space-y-4'>
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className='h-20 bg-stone-100 dark:bg-stone-800 rounded-lg animate-pulse'
            />
          ))}
        </div>
      ) : articles.length === 0 ? (
        <div className='text-center py-12'>
          <FiBookmark
            size={40}
            className='mx-auto text-stone-300 dark:text-stone-600 mb-4'
          />
          <p className='text-stone-400 text-lg'>保存した記事はまだありません</p>
          <p className='text-stone-400 text-sm mt-1'>
            記事のブックマークアイコンをクリックして保存できます
          </p>
        </div>
      ) : (
        <div className='space-y-3'>
          {articles.map((article) => (
            <Link
              key={article.id}
              href={`/wiki/${article.id}`}
              className='block p-4 rounded-lg border border-stone-200 dark:border-stone-700 hover:border-stone-400 dark:hover:border-stone-500 transition-colors'
            >
              <h2 className='font-semibold text-lg mb-1'>{article.title}</h2>
              <p className='text-stone-400 text-sm'>
                {new Date(article.created_at.seconds * 1000).toLocaleDateString(
                  'ja-JP',
                )}
              </p>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
