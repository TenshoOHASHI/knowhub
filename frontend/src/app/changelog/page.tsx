'use client';

import { useEffect, useState } from 'react';
import { FiCalendar, FiChevronRight, FiChevronDown } from 'react-icons/fi';
import { motion, AnimatePresence } from 'motion/react';

interface Update {
  date: string;
  title: string;
  changes: string[];
}

export default function ChangelogPage() {
  const [updates, setUpdates] = useState<Update[]>([]);
  const [expanded, setExpanded] = useState<string | null>(null);

  // データを取得、保存、再レンダリング
  useEffect(() => {
    fetch('/api/changelog')
      .then((res) => res.json())
      .then((data) => setUpdates(data))
      .catch(() => {});
  }, []);

  const toggle = (key: string) => {
    // もし今開いているキーとクリックされたキーが同じなら、閉じる
    // キーが更新され、再描画
    setExpanded(expanded === key ? null : key);
  };

  return (
    <div className='h-full overflow-y-auto'>
      <div className='max-w-3xl mx-auto px-6 py-8'>
        <h1 className='text-4xl font-bold mb-2'>Changelog</h1>
        <p className='text-md text-gray-500 dark:text-stone-400 mb-8'>
          TenHub の更新履歴
        </p>

        <div className='space-y-3'>
          {/* データ抽出 */}
          {updates.map((update) => {
            // データからキーを作成
            const key = update.date + update.title;
            // 初回はnull値のためfalse、再描画時に、expandedが更新され、trueになる
            const isOpen = expanded === key;

            return (
              <motion.div
                key={key}
                className='rounded-xl border border-stone-200 dark:border-stone-700/50 bg-white dark:bg-stone-800/30 overflow-hidden'
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.3 }}
              >
                {/* Header */}
                <button
                  onClick={() => toggle(key)}
                  className='w-full flex items-center gap-3 px-5 py-4 text-left hover:bg-stone-50 dark:hover:bg-stone-800/50 transition-colors'
                >
                  <FiCalendar className='text-stone-400 shrink-0' size={16} />
                  <span className='text-xs text-stone-400 shrink-0'>
                    {update.date}
                  </span>
                  <span className='text-sm font-semibold flex-1'>
                    {update.title}
                  </span>
                  <span className='text-xs text-stone-400'>
                    {update.changes.length} items
                  </span>
                  {isOpen ? (
                    <FiChevronDown className='text-stone-400' size={16} />
                  ) : (
                    <FiChevronRight className='text-stone-400' size={16} />
                  )}
                </button>

                {/* Changes */}
                <AnimatePresence>
                  {isOpen && (
                    <motion.div
                      initial={{ height: 0, opacity: 0 }}
                      animate={{ height: 'auto', opacity: 1 }}
                      exit={{ height: 0, opacity: 0 }}
                      transition={{ duration: 0.25 }}
                      className='overflow-hidden'
                    >
                      <div className='px-5 pb-4 pt-1 border-t border-stone-100 dark:border-stone-700/50'>
                        <ul className='space-y-2'>
                          {/* 配列からデータを抽出 */}
                          {update.changes.map((change, i) => (
                            <motion.li
                              key={i}
                              className='flex items-start gap-2 text-sm text-gray-600 dark:text-stone-400'
                              initial={{ opacity: 0, x: -10 }}
                              animate={{ opacity: 1, x: 0 }}
                              transition={{ delay: i * 0.05 }}
                            >
                              <FiChevronRight
                                className='text-red-400 shrink-0 mt-0.5'
                                size={14}
                              />
                              {change}
                            </motion.li>
                          ))}
                        </ul>
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              </motion.div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
