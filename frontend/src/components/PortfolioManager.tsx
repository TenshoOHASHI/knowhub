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
  FiChevronLeft,
  FiChevronRight,
} from 'react-icons/fi';
import { useEffect, useState, useMemo } from 'react';
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

  const load = () => {
    getPortfolioItems()
      .then((data) => {
        setItems(data.items || []);
      })
      .catch(() => toast('取得に失敗しました', 'error'))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    load();
  }, []);

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
      toast(errorMsg);
      setMessage(errorMsg);
    }
  };

  const handleDelete = async (id: string) => {
    if (!deleteTarget) return;

    try {
      await deletePortfolioItem(id);
      toast('削除しました');
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
    return <div className='animate-pulse text-gray-400 p-6'>Loading...</div>;
  }

  return (
    <div className='max-w-4xl mx-auto space-y-6 p-4'>
      {/* Form */}
      <form onSubmit={handleSubmit} className='space-y-4'>
        <h2 className='text-3xl font-semibold'>
          {editId ? 'ポートフォリオ編集' : 'ポートフォリオ作成'}
        </h2>

        <div>
          <label className='block mb-1 text-sm'>タイトル</label>
          <input
            type='text'
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder='プロジェクト名'
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 bg-transparent'
          />
        </div>

        <div>
          <label className='block mb-1 text-sm'>説明</label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder='プロジェクトの説明'
            rows={3}
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 bg-transparent'
          />
        </div>

        <div>
          <label className='block mb-1 text-sm'>URL</label>
          <input
            type='text'
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder='https://github.com/user/repo'
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 bg-transparent'
          />
        </div>

        <div>
          <label className='block mb-1 text-sm'>ステータス</label>
          <select
            value={status}
            onChange={(e) =>
              setStatus(e.target.value as 'developing' | 'completed')
            }
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 bg-transparent'
          >
            <option value='developing'>開発中</option>
            <option value='completed'>完了</option>
          </select>
        </div>

        <div>
          <label className='block mb-1 text-sm'>カテゴリー</label>
          <select
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 bg-transparent'
          >
            <option value='project'>プロジェクト</option>
            <option value='assignment'>課題</option>
            <option value='contribution'>貢献</option>
          </select>
        </div>

        <div>
          <label className='block mb-1 text-sm'>
            Tech Stack（カンマ区切り）
          </label>
          <input
            type='text'
            value={techStackInput}
            onChange={(e) => setTechStackInput(e.target.value)}
            placeholder='Go, React, MySQL, Docker'
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 bg-transparent'
          />
        </div>

        <div className='flex gap-2'>
          <button
            type='submit'
            className='flex items-center gap-1 px-4 py-2 rounded-lg bg-green-700 dark:bg-green-300 text-white dark:text-stone-900 hover:bg-green-800 dark:hover:bg-green-300/90'
          >
            {editId ? <FiCheck /> : <FiPlus />}
            {editId ? '更新' : '作成'}
          </button>

          {editId && (
            <button
              type='button'
              onClick={resetForm}
              className='flex items-center gap-1 px-4 py-2 rounded-lg border border-stone-300 dark:border-stone-600 hover:bg-stone-100 dark:hover:bg-stone-800'
            >
              <FiX />
              キャンセル
            </button>
          )}
        </div>
        {message &&
          (message.includes('更新しました') ||
            message.includes('作成しました')) && (
            <Link
              href='/portfolio'
              className='underline hover:text-green-800 dark:hover:text-green-300'
            >
              ポートフォリオを確認 →
            </Link>
          )}
      </form>

      {/* Item List (paged slider) */}
      <div className='space-y-2'>
        <div className='flex items-center justify-between'>
          <h2 className='text-lg font-semibold'>ポートフォリオ一覧</h2>
          {totalPages > 1 && (
            <span className='text-xs text-gray-400'>
              {page + 1} / {totalPages}
            </span>
          )}
        </div>

        {items.length === 0 && (
          <p className='text-sm text-gray-400'>アイテムがありません</p>
        )}

        <div className='space-y-2'>
          {pageItems.map((item) => (
            <div
              key={item.id}
              className='flex items-center justify-between gap-3 p-3 rounded-lg border border-stone-200 dark:border-stone-700'
            >
              <div className='flex-1 min-w-0'>
                <p className='font-medium truncate'>{item.title}</p>
                <p className='text-xs text-gray-400'>
                  {item.status === 'completed' ? '完了' : '開発中'}
                  {item.category && ` · ${categoryLabel(item.category)}`}
                </p>
              </div>
              <div className='flex items-center gap-1'>
                <button
                  onClick={() => startEdit(item)}
                  className='p-2 rounded hover:bg-stone-100 dark:hover:bg-stone-800 text-gray-500'
                >
                  <FiEdit2 size={16} />
                </button>
                <button
                  onClick={() => setDeleteTarget(item.id)}
                  className='p-2 rounded hover:bg-red-50 dark:hover:bg-red-900/20 text-gray-500 hover:text-red-500'
                >
                  <FiTrash2 size={16} />
                </button>
                {deleteTarget && (
                  <ConfirmModal
                    isOpen={true}
                    title='ポートフォリオ削除'
                    message={`${item.title}を削除しますか？`}
                    onConfirm={() => handleDelete(item.id)}
                    onCancel={() => setDeleteTarget(null)}
                  />
                )}
              </div>
            </div>
          ))}
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className='flex items-center justify-center gap-3 pt-2'>
            <button
              onClick={() => setPage((p) => Math.max(0, p - 1))}
              disabled={page === 0}
              className='p-2 rounded-lg border border-stone-200 dark:border-stone-700 hover:bg-stone-100 dark:hover:bg-stone-800 disabled:opacity-30 disabled:cursor-not-allowed transition-colors'
            >
              <FiChevronLeft size={16} />
            </button>
            <div className='flex gap-1'>
              {Array.from({ length: totalPages }, (_, i) => (
                <button
                  key={i}
                  onClick={() => setPage(i)}
                  className={`w-2 h-2 rounded-full transition-colors ${
                    i === page
                      ? 'bg-green-700 dark:bg-green-300'
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
              <FiChevronRight size={16} />
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
