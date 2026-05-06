'use client';

import { getPortfolioItems } from '@/lib/api';
import { type PortfolioItem } from '@/lib/types';
import { motion } from 'motion/react';
import { useEffect, useState } from 'react';
import { FiExternalLink } from 'react-icons/fi';

export default function PortfolioPage() {
  const [items, setItems] = useState<PortfolioItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getPortfolioItems()
      .then((data) => {
        setItems(data.items || []);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className='flex items-center justify-center min-h-[60vh]'>
        <div className='animate-pulse text-gray-400'>Loading...</div>
      </div>
    );
  }

  return (
    <div className='h-full overflow-y-auto'>
      <div className='max-w-4xl mx-auto px-6 py-8'>
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className='mb-8'
        >
          <h1 className='text-3xl font-bold mb-1'>ポートフォリオ</h1>
          <p className='text-md text-gray-500 dark:text-stone-400'>
            これまでに手がけたプロジェクト
          </p>
        </motion.div>

        {items.length === 0 && (
          <p className='text-stone-400 text-center py-20'>
            No portfolio items yet
          </p>
        )}

        {/* Project List */}
        <div className='space-y-4'>
          {items.map((item, i) => {
            const isCompleted = item.status === 'completed';
            let techStack: string[] = [];
            try {
              const parsed = JSON.parse(item.tech_stack || '[]');
              if (Array.isArray(parsed)) techStack = parsed;
            } catch {}

            return (
              <motion.article
                key={item.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.06, duration: 0.3 }}
                className='group border-b border-stone-200 dark:border-stone-800 pb-6'
              >
                <div className='flex items-start justify-between gap-4'>
                  <div className='flex-1 min-w-0'>
                    {/* Title row */}
                    <div className='flex items-center gap-2 mb-1'>
                      <h2 className='text-lg font-semibold truncate'>
                        {item.title}
                      </h2>
                      <span
                        className={`text-xs px-2 py-0.5 rounded-full shrink-0 ${
                          isCompleted
                            ? 'text-stone-500 bg-stone-100 dark:bg-stone-800 dark:text-stone-400'
                            : 'text-stone-500 bg-stone-100 dark:bg-stone-800 dark:text-stone-400'
                        }`}
                      >
                        {isCompleted ? 'DONE' : 'WIP'}
                      </span>
                    </div>

                    {/* Description */}
                    <p className='text-sm text-gray-500 dark:text-stone-400 line-clamp-2 mb-2'>
                      {item.description}
                    </p>

                    {/* Tech stack */}
                    {techStack.length > 0 && (
                      <p className='text-xs text-stone-400 dark:text-stone-500 font-mono mb-1'>
                        {techStack.join(' / ')}
                      </p>
                    )}

                    {/* Date */}
                    {item.created_at?.seconds && (
                      <p className='text-xs text-stone-400 dark:text-stone-500'>
                        {new Date(item.created_at.seconds * 1000).toLocaleDateString('ja-JP', {
                          year: 'numeric',
                          month: 'short',
                          day: 'numeric',
                        })}
                      </p>
                    )}
                  </div>

                  {/* Link */}
                  {item.url && (
                    <a
                      href={item.url}
                      target='_blank'
                      rel='noopener noreferrer'
                      className='shrink-0 p-2 text-stone-400 hover:text-black dark:hover:text-white transition-colors'
                    >
                      <FiExternalLink size={18} />
                    </a>
                  )}
                </div>
              </motion.article>
            );
          })}
        </div>
      </div>
    </div>
  );
}
