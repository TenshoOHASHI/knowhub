'use client';

import { useToast } from '@/context/ToastContext';
import { type PortfolioItem } from '@/lib/types';
import {
  getPortfolioItems,
  savePortfolioItem,
  deletePortfolioItem,
} from '@/lib/api';
import {
  FiPlus,
  FiTrash2,
  FiEdit2,
  FiX,
  FiCheck,
  FiBriefcase,
  FiChevronLeft,
  FiChevronRight,
} from 'react-icons/fi';
import { useEffect, useState, useMemo, useCallback } from 'react';
import Link from 'next/link';
import ConfirmModal from './ConfirmModal';

const categoryLabels: Record<string, string> = {
  project: 'プロジェクト',
  assignment: '課題',
  contribution: '貢献',
};

function categoryLabel(cat: string) {
  return categoryLabels[cat] || cat;
}

export function PortfolioManager() {
  const [items, setItems] = useState<PortfolioItem[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  // Form state
  const [editId, setEditId] = useState<string | null>(null);
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [url, setUrl] = useState('');
  const [status, setStatus] = useState<'developing' | 'completed'>(
    'developing',
  );
  const [category, setCategory] = useState('project');
  const [techStackInput, setTechStackInput] = useState('');
  const [message, setMessage] = useState('');
  const [page, setPage] = useState(0);
  const PER_PAGE = 3;

  const totalPages = Math.ceil(items.length / PER_PAGE);
  const pageItems = useMemo(
    () => items.slice(page * PER_PAGE, page * PER_PAGE + PER_PAGE),
    [items, page],
  );

  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const load = useCallback(() => {
    getPortfolioItems()
      .then((data) => {
        setItems(data.items || []);
      })
      .catch(() => toast('取得に失敗しました', 'error'))
      .finally(() => setLoading(false));
  }, [toast]);

  useEffect(() => {
    load();
  }, [load]);

  const resetForm = () => {
    setEditId(null);
    setTitle('');
    setDescription('');
    setUrl('');
    setStatus('developing');
    setCategory('project');
    setTechStackInput('');
  };

  const handleSubmit = async (e: React.SubmitEvent) => {
    e.preventDefault();
    if (!title.trim() || !description.trim()) {
      toast('タイトルと説明は必須です', 'error');
      return;
    }

    try {
      const techStack = JSON.stringify(
        techStackInput
          .split(',')
          .map((s) => s.trim())
          .filter(Boolean),
      );

      await savePortfolioItem({
        id: editId || undefined,
        title,
        description,
        url,
        status,
        category,
        tech_stack: techStack,
      });
      const successMsg = editId ? '更新しました' : '作成しました';
      toast(successMsg);
      setMessage(successMsg);
      resetForm();
      load();
    } catch {
      const errorMsg = editId ? '更新に失敗しました' : '作成に失敗しました';
      toast(errorMsg, 'error');
      setMessage(errorMsg);
    }
  };

  const handleDelete = async (id: string) => {
    if (!deleteTarget) return;

    try {
      await deletePortfolioItem(id);
      toast('削除しました', 'success');
      setDeleteTarget(null);
      load();
    } catch {
      toast('削除に失敗しました', 'error');
    }
  };

  const startEdit = (item: PortfolioItem) => {
    setEditId(item.id);
    setTitle(item.title);
    setDescription(item.description);
    setUrl(item.url);
    setStatus(item.status);
    setCategory(item.category || 'project');
    try {
      const parsed = JSON.parse(item.tech_stack || '[]');
      setTechStackInput(Array.isArray(parsed) ? parsed.join(', ') : '');
    } catch {
      setTechStackInput('');
    }
  };

  if (loading) {
    return <div className='animate-pulse text-stone-400 p-6'>Loading...</div>;
  }

  return (
    <div className='p-6 space-y-6'>
      {/* ヘッダー */}
      <div>
        <h2 className='text-lg font-semibold text-stone-900 dark:text-stone-100 mb-1'>ポートフォリオ管理</h2>
        <p className='text-sm text-stone-500 dark:text-stone-400'>プロジェクトや実績を管理します</p>
      </div>

      {/* フォーム + 一覧の2カラムレイアウト */}
      <div className='grid grid-cols-1 lg:grid-cols-2 gap-6'>
        {/* 作成・編集フォーム */}
        <div className='space-y-4'>
          <h3 className='text-sm font-semibold text-stone-700 dark:text-stone-300'>
            {editId ? '編集' : '新規作成'}
          </h3>

          <form onSubmit={handleSubmit} className='space-y-3'>
            <div>
              <label className='block text-xs font-medium mb-1 text-stone-600 dark:text-stone-400'>タイトル</label>
              <input
                type='text'
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder='プロジェクト名'
                className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-3 py-2 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-stone-500 transition'
              />
            </div>

            <div>
              <label className='block text-xs font-medium mb-1 text-stone-600 dark:text-stone-400'>説明</label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder='プロジェクトの説明'
                rows={3}
                className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-3 py-2 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-stone-500 transition resize-none'
              />
            </div>

            <div>
              <label className='block text-xs font-medium mb-1 text-stone-600 dark:text-stone-400'>URL</label>
              <input
                type='text'
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                placeholder='https://github.com/user/repo'
                className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-3 py-2 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-stone-500 transition'
              />
            </div>

            <div className='grid grid-cols-2 gap-3'>
              <div>
                <label className='block text-xs font-medium mb-1 text-stone-600 dark:text-stone-400'>ステータス</label>
                <select
                  value={status}
                  onChange={(e) => setStatus(e.target.value as 'developing' | 'completed')}
                  className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-3 py-2 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-stone-500 transition'
                >
                  <option value='developing'>開発中</option>
                  <option value='completed'>完了</option>
                </select>
              </div>

              <div>
                <label className='block text-xs font-medium mb-1 text-stone-600 dark:text-stone-400'>カテゴリ</label>
                <select
                  value={category}
                  onChange={(e) => setCategory(e.target.value)}
                  className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-3 py-2 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-stone-500 transition'
                >
                  <option value='project'>プロジェクト</option>
                  <option value='assignment'>課題</option>
                  <option value='contribution'>貢献</option>
                </select>
              </div>
            </div>

            <div>
              <label className='block text-xs font-medium mb-1 text-stone-600 dark:text-stone-400'>Tech Stack（カンマ区切り）</label>
              <input
                type='text'
                value={techStackInput}
                onChange={(e) => setTechStackInput(e.target.value)}
                placeholder='Go, React, MySQL, Docker'
                className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-3 py-2 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-stone-500 transition'
              />
            </div>

            <div className='flex gap-2'>
              <button
                type='submit'
                className='flex-1 bg-stone-900 text-white dark:bg-stone-100 dark:text-stone-900 px-4 py-2 rounded-lg text-sm font-medium hover:bg-stone-800 dark:hover:bg-stone-200 transition-colors flex items-center justify-center gap-2'
              >
                {editId ? <FiCheck /> : <FiPlus />}
                {editId ? '更新' : '作成'}
              </button>

              {editId && (
                <button
                  type='button'
                  onClick={resetForm}
                  className='px-4 py-2 rounded-lg border border-stone-300 dark:border-stone-600 hover:bg-stone-100 dark:hover:bg-stone-800 text-stone-600 dark:text-stone-400 text-sm transition-colors'
                >
                  <FiX />
                </button>
              )}
            </div>

            {message &&
              (message.includes('更新しました') ||
                message.includes('作成しました')) && (
                <Link
                  href='/portfolio'
                  className='text-xs text-stone-600 dark:text-stone-400 underline hover:text-stone-900 dark:hover:text-stone-200 flex items-center gap-1'
                >
                  <FiCheck size={12} />
                  Portfolioで確認 →
                </Link>
              )}
          </form>
        </div>

        {/* 一覧 */}
        <div className='space-y-3'>
          <div className='flex items-center justify-between'>
            <h3 className='text-sm font-semibold text-stone-700 dark:text-stone-300'>
              ポートフォリオ一覧 ({items.length})
            </h3>
            {totalPages > 1 && (
              <span className='text-xs text-stone-400'>
                {page + 1} / {totalPages}
              </span>
            )}
          </div>

          {items.length === 0 ? (
            <p className='text-stone-400 text-sm text-center py-8'>アイテムがありません</p>
          ) : (
            <>
              <div className='space-y-2'>
                {pageItems.map((item) => (
                  <div
                    key={item.id}
                    className={`p-4 rounded-lg border transition-all cursor-pointer hover:shadow-sm ${
                      editId === item.id
                        ? 'bg-stone-100 dark:bg-stone-800 border-stone-400 dark:border-stone-500'
                        : 'bg-white dark:bg-stone-900 border-stone-200 dark:border-stone-700 hover:border-stone-300 dark:hover:border-stone-600'
                    }`}
                    onClick={() => startEdit(item)}
                  >
                    <div className='flex items-start justify-between gap-2'>
                      <div className='flex-1 min-w-0'>
                        <p className='font-medium text-sm text-stone-900 dark:text-stone-100 truncate'>
                          {item.title}
                        </p>
                        <div className='flex items-center gap-2 mt-1'>
                          <span className={`text-xs px-2 py-0.5 rounded-full ${
                            item.status === 'completed'
                              ? 'bg-stone-800 text-white dark:bg-stone-200 dark:text-stone-900'
                              : 'bg-stone-200 text-stone-700 dark:bg-stone-700 dark:text-stone-300'
                          }`}>
                            {item.status === 'completed' ? '完了' : '開発中'}
                          </span>
                          {item.category && (
                            <span className='text-xs text-stone-500 dark:text-stone-400'>
                              · {categoryLabel(item.category)}
                            </span>
                          )}
                        </div>
                      </div>
                      <div className='flex items-center gap-1'>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            setDeleteTarget(item.id);
                          }}
                          className='p-1.5 rounded hover:bg-stone-100 dark:hover:bg-stone-800 text-stone-400 hover:text-stone-600 dark:hover:text-stone-300 transition-colors'
                          title='削除'
                        >
                          <FiTrash2 size={14} />
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>

              {/* ペジネーション */}
              {totalPages > 1 && (
                <div className='flex items-center justify-center gap-2 pt-2'>
                  <button
                    onClick={() => setPage((p) => Math.max(0, p - 1))}
                    disabled={page === 0}
                    className='p-2 rounded-lg border border-stone-200 dark:border-stone-700 hover:bg-stone-100 dark:hover:bg-stone-800 disabled:opacity-30 disabled:cursor-not-allowed transition-colors'
                  >
                    <FiChevronLeft size={14} />
                  </button>
                  <div className='flex gap-1'>
                    {Array.from({ length: totalPages }, (_, i) => (
                      <button
                        key={i}
                        onClick={() => setPage(i)}
                        className={`w-2 h-2 rounded-full transition-colors ${
                          i === page
                            ? 'bg-stone-800 dark:bg-stone-200'
                            : 'bg-stone-300 dark:bg-stone-600'
                        }`}
                      />
                    ))}
                  </div>
                  <button
                    onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
                    disabled={page === totalPages - 1}
                    className='p-2 rounded-lg border border-stone-200 dark:border-stone-700 hover:bg-stone-100 dark:hover:bg-stone-800 disabled:opacity-30 disabled:cursor-not-allowed transition-colors'
                  >
                    <FiChevronRight size={14} />
                  </button>
                </div>
              )}
            </>
          )}
        </div>
      </div>

      {deleteTarget && (
        <ConfirmModal
          isOpen={true}
          title='ポートフォリオ削除'
          message={`${items.find(i => i.id === deleteTarget)?.title || 'このアイテム'}を削除しますか？`}
          onConfirm={async () => {
            await handleDelete(deleteTarget);
          }}
          onCancel={() => setDeleteTarget(null)}
        />
      )}
    </div>
  );
}
