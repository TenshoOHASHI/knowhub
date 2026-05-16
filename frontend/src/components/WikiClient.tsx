'use client';
import { useState, useEffect, useRef } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { FiSearch, FiCalendar, FiFileText, FiLock, FiX, FiMapPin } from 'react-icons/fi';
import type { Article, Category } from '@/lib/types';
import { getCategories } from '@/lib/api';
import { motion } from 'motion/react';

/** 指定カテゴリIDとその全子孫カテゴリIDのセットを返す */
function getDescendantIds(categoryId: string, categories: Category[]): Set<string> {
  const childrenMap = new Map<string, string[]>();
  for (const c of categories) {
    if (c.parent_id) {
      const list = childrenMap.get(c.parent_id) || [];
      list.push(c.id);
      childrenMap.set(c.parent_id, list);
    }
  }
  const ids = new Set<string>([categoryId]);
  const queue = [categoryId];
  while (queue.length > 0) {
    const current = queue.pop()!;
    const children = childrenMap.get(current) || [];
    for (const child of children) {
      if (!ids.has(child)) {
        ids.add(child);
        queue.push(child);
      }
    }
  }
  return ids;
}

// Markdown記号を除去してプレーンテキストにする
function stripMarkdown(md: string): string {
  return md
    .replace(/```[\s\S]*?```/g, '') // コードブロック除去
    .replace(/<details>[\s\S]*?<\/details>/gi, '') // 折りたたみブロック
    .replace(/<[^>]+>/g, '') // HTMLタグ除去
    .replace(/^:::message.*$\n?/gm, '') // Zennコールアウト開始
    .replace(/^:::\s*$/gm, '') // Zennコールアウト終了
    .replace(/`([^`]+)`/g, '$1') // インラインコード
    .replace(/!\[.*?\]\(.*?\)/g, '') // 画像
    .replace(/\[([^\]]+)\]\(.*?\)/g, '$1') // リンク → テキストだけ
    .replace(/#{1,6}\s/g, '') // 見出し記号
    .replace(/(\*{1,3}|_{1,3})(.*?)\1/g, '$2') // 太字・斜体
    .replace(/^>\s.*/gm, '') // 引用
    .replace(/^[-*+]\s/gm, '') // リスト記号
    .replace(/^\|.+\|$/gm, '') // テーブル行（| ... |）
    .replace(/^---+$/gm, '') // 水平線
    .replace(/^===+$/gm, '') // setext見出し下線
    .replace(/\[!(NOTE|INFO|TIP|WARNING|CAUTION|IMPORTANT)\]/g, '') // GitHub callout marker
    .replace(/\n{2,}/g, ' ') // 改行をスペースに
    .trim();
}

export default function WikiClient({ articles }: { articles: Article[] }) {
  const [query, setQuery] = useState('');
  const [searchDialogOpen, setSearchDialogOpen] = useState(false);
  const [searchDialogQuery, setSearchDialogQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const [categories, setCategories] = useState<Category[]>([]);

  const searchParams = useSearchParams();
  const categoryId = searchParams.get('category');

  useEffect(() => {
    getCategories()
      .then((data) => setCategories(data.categories || []))
      .catch(() => {});
  }, []);

  // ダイアログ用のフィルタリング
  const dialogFiltered = articles.filter((a) => {
    const q = searchDialogQuery.toLowerCase();
    return (
      a.title.toLowerCase().includes(q) || a.content.toLowerCase().includes(q)
    );
  });

  // 選択カテゴリの子孫IDセットを構築
  const categoryIds = categoryId ? getDescendantIds(categoryId, categories) : null;

  // 既存の検索バー用のフィルタリング + ピン留め記事を先頭に
  const filtered = articles
    .filter((a) => {
      const q = query.toLowerCase();
      const matchesQuery =
        a.title.toLowerCase().includes(q) || a.content.toLowerCase().includes(q);
      const matchesCategory = !categoryIds || categoryIds.has(a.category_id);
      return matchesQuery && matchesCategory;
    })
    .sort((a, b) => {
      if (a.is_pinned && !b.is_pinned) return -1;
      if (!a.is_pinned && b.is_pinned) return 1;
      return 0;
    });

  // 「/」キーで検索ダイアログを開く
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // 入力欄にフォーカスがある場合は無視
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement
      ) {
        return;
      }

      // 「/」キーで検索ダイアログを開く
      if (e.key === '/') {
        e.preventDefault();
        setSearchDialogOpen(true);
        setSearchDialogQuery('');
        setSelectedIndex(0);
      }

      // Cmd/Ctrl + K でも開く
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setSearchDialogOpen(true);
        setSearchDialogQuery('');
        setSelectedIndex(0);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  // ダイアログ内のキーボードナビゲーション
  useEffect(() => {
    if (!searchDialogOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setSearchDialogOpen(false);
        return;
      }

      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setSelectedIndex((i) => Math.min(i + 1, dialogFiltered.length - 1));
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setSelectedIndex((i) => Math.max(i - 1, 0));
      } else if (e.key === 'Enter' && dialogFiltered.length > 0) {
        e.preventDefault();
        const selected = dialogFiltered[selectedIndex];
        if (selected) {
          window.location.href = `/wiki/${selected.id}`;
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [searchDialogOpen, dialogFiltered, selectedIndex]);

  // ダイアログが開いたときにフォーカス
  useEffect(() => {
    if (searchDialogOpen) {
      setTimeout(() => {
        searchInputRef.current?.focus();
      }, 100);
    }
  }, [searchDialogOpen]);

  // 選択項目がスクロール範囲外の場合はスクロール
  useEffect(() => {
    if (listRef.current && dialogFiltered.length > 0) {
      const items = listRef.current.querySelectorAll('[data-index]');
      const selectedEl = items[selectedIndex] as HTMLElement;
      if (selectedEl) {
        selectedEl.scrollIntoView({ block: 'nearest' });
      }
    }
  }, [selectedIndex, dialogFiltered.length]);

  return (
    <>
      <div className='max-w-4xl mx-auto p-6'>
        {/* 検索バー */}
        <div className='relative mb-6'>
          <FiSearch className='absolute left-3 top-1/2 -translate-y-1/2 text-gray-400' />
          <input
            type='text'
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder='記事を検索...'
            className='w-full border border-black dark:border-stone-600 rounded-lg pl-10 pr-16 py-2 bg-transparent'
          />
          <span className='absolute right-3 top-1/2 -translate-y-1/2 text-xs text-gray-400 select-none'>
            <kbd className='px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded border border-gray-300 dark:border-gray-600 font-mono'>/</kbd>
          </span>
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
              Knowledge Base
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
              className={`border rounded-lg p-5
              hover:shadow-md transition-all duration-200 overflow-hidden relative ${
                a.is_pinned
                  ? 'border-amber-300 dark:border-amber-700 bg-amber-50/30 dark:bg-amber-900/10 hover:border-amber-400 dark:hover:border-amber-600'
                  : 'border-black/10 dark:border-stone-700 hover:border-black/20 dark:hover:border-stone-500'
              }`}
            >
              {a.is_pinned && (
                <span className='absolute top-3 right-3 inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-amber-100 dark:bg-amber-900/50 text-amber-600 dark:text-amber-400 text-xs font-medium'>
                  <FiMapPin size={10} />
                  TOP
                </span>
              )}
              <h2 className='text-lg font-semibold group-hover:text-red-500/70 transition-colors flex items-center gap-2'>
                {a.visibility === 'locked' && (
                  <span className='inline-flex items-center justify-center w-5 h-5 rounded bg-stone-200 dark:bg-stone-700 shrink-0'>
                    <FiLock
                      size={11}
                      className='text-stone-500 dark:text-stone-400'
                    />
                  </span>
                )}
                <span className='min-w-0 break-all text-2xl'>{a.title}</span>
              </h2>
              <p className='text-gray-500 dark:text-stone-400 mt-2 line-clamp-2 text-md leading-relaxed'>
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

      {/* クイック検索ダイアログ */}
      {searchDialogOpen && (
        <div
          className='fixed inset-0 z-50 bg-black/40 backdrop-blur-sm flex items-start justify-center pt-[20vh] p-4'
          onClick={() => setSearchDialogOpen(false)}
        >
          <div
            className='bg-white dark:bg-stone-900 rounded-xl shadow-2xl w-full max-w-2xl overflow-hidden flex flex-col max-h-[60vh]'
            onClick={(e) => e.stopPropagation()}
          >
            {/* 検索入力欄 */}
            <div className='flex items-center gap-3 px-4 py-3 border-b border-stone-200 dark:border-stone-700 bg-stone-50 dark:bg-stone-800/50'>
              <FiSearch className='text-stone-400 shrink-0' />
              <input
                ref={searchInputRef}
                type='text'
                value={searchDialogQuery}
                onChange={(e) => {
                  setSearchDialogQuery(e.target.value);
                  setSelectedIndex(0);
                }}
                placeholder='記事を検索... (↑↓で選択、Enterで開く)'
                className='flex-1 bg-transparent border-none outline-none text-stone-900 dark:text-stone-100 placeholder-stone-400'
              />
              <button
                onClick={() => setSearchDialogOpen(false)}
                className='text-stone-400 hover:text-stone-600 dark:hover:text-stone-200 transition-colors p-1 rounded hover:bg-stone-200 dark:hover:bg-stone-700'
                title='閉じる (ESC)'
              >
                <FiX size={18} />
              </button>
            </div>

            {/* 記事リスト */}
            <div ref={listRef} className='flex-1 overflow-y-auto thin-scrollbar'>
              {dialogFiltered.length === 0 ? (
                <div className='flex flex-col items-center justify-center py-12 text-stone-400'>
                  <FiFileText size={40} className='mb-2' />
                  <p>
                    {searchDialogQuery
                      ? '一致する記事が見つかりません'
                      : '記事が存在しません'}
                  </p>
                </div>
              ) : (
                <div className='py-1'>
                  {dialogFiltered.map((article, index) => (
                    <Link
                      key={article.id}
                      href={`/wiki/${article.id}`}
                      data-index={index}
                      onClick={() => setSearchDialogOpen(false)}
                      onMouseEnter={() => setSelectedIndex(index)}
                      className={`block w-full text-left px-4 py-3 border-l-2 transition-colors ${
                        index === selectedIndex
                          ? 'bg-blue-50 dark:bg-blue-900/20 border-blue-500'
                          : 'bg-transparent border-transparent hover:bg-stone-50 dark:hover:bg-stone-800/30'
                      }`}
                    >
                      <div className='flex items-start gap-3'>
                        <div className='flex-1 min-w-0'>
                          <div className='flex items-center gap-2 mb-1'>
                            {article.visibility === 'locked' && (
                              <FiLock
                                size={12}
                                className='text-stone-400 shrink-0'
                              />
                            )}
                            <h3
                              className={`font-medium truncate ${
                                index === selectedIndex
                                  ? 'text-blue-700 dark:text-blue-300'
                                  : 'text-stone-900 dark:text-stone-100'
                              }`}
                            >
                              {article.title}
                            </h3>
                          </div>
                          <p className='text-xs text-stone-500 dark:text-stone-400 line-clamp-1'>
                            {article.visibility === 'locked'
                              ? 'この記事は限定公開です'
                              : stripMarkdown(article.content).slice(0, 100)}
                          </p>
                          <div className='flex items-center gap-1.5 mt-1.5 text-xs text-stone-400 dark:text-stone-500'>
                            <FiCalendar size={10} />
                            <span>
                              {article.created_at
                                ? new Date(
                                    article.created_at.seconds * 1000,
                                  ).toLocaleDateString('ja-JP')
                                : ''}
                            </span>
                          </div>
                        </div>
                        {index === selectedIndex && (
                          <span className='text-xs text-stone-400 dark:text-stone-500 self-center'>
                            Enter
                          </span>
                        )}
                      </div>
                    </Link>
                  ))}
                </div>
              )}
            </div>

            {/* フッター */}
            <div className='px-4 py-2 border-t border-stone-200 dark:border-stone-700 bg-stone-50 dark:bg-stone-800/50 text-xs text-stone-500 dark:text-stone-400 flex items-center justify-between'>
              <span>{dialogFiltered.length} 件の記事</span>
              <div className='flex gap-3'>
                <span>↑↓ 選択</span>
                <span>Enter 開く</span>
                <span>ESC 閉じる</span>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
