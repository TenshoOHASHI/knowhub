'use client';

import { useState, useEffect, useCallback } from 'react';
import { useToast } from '@/context/ToastContext';
import type { Category } from '@/lib/types';
import { getCategories, createCategory, deleteCategory } from '@/lib/api';
import ConfirmModal from './ConfirmModal';
import { FiFile, FiFolder, FiTrash2 } from 'react-icons/fi';

export function CategoryManager() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [name, setName] = useState('');
  const [parentId, setParentId] = useState('');
  const { toast } = useToast();
  const [deleteTarget, setDeleteTarget] = useState<Category | null>(null);

  const load = useCallback(() => {
    getCategories()
      .then((data) => setCategories(data.categories || []))
      .catch(() => toast('取得に失敗しました', 'error'));
  }, [toast]);

  useEffect(() => {
    load();
  }, [load]);

  const handleCreate = async (e: React.SubmitEvent) => {
    e.preventDefault();
    try {
      await createCategory({ name, parent_id: parentId });
      toast('カテゴリを作成しました', 'success');
      setName('');
      setParentId('');
      load();
    } catch {
      toast('作成に失敗しました', 'error');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteCategory(id);
      toast('カテゴリを削除しました', 'success');
      setDeleteTarget(null);
      load();
    } catch {
      toast('削除に失敗しました', 'error');
    }
  };

  return (
    <div className='p-6 space-y-6'>
      {/* ヘッダー */}
      <div>
        <h2 className='text-lg font-semibold text-stone-900 dark:text-stone-100 mb-1'>カテゴリ管理</h2>
        <p className='text-sm text-stone-500 dark:text-stone-400'>Wikiのカテゴリを作成・管理します</p>
      </div>

      {/* 作成フォーム */}
      <form onSubmit={handleCreate} className='flex gap-3 items-end'>
        <div className='flex-1'>
          <label className='block text-sm font-medium mb-1.5 text-stone-700 dark:text-stone-300'>カテゴリ名</label>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-3 py-2 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-stone-500 transition'
            placeholder='例: Go'
            required
          />
        </div>
        <div className='flex-1'>
          <label className='block text-sm font-medium mb-1.5 text-stone-700 dark:text-stone-300'>親カテゴリ</label>
          <select
            value={parentId}
            onChange={(e) => setParentId(e.target.value)}
            className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-3 py-2 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-stone-500 transition'
          >
            <option value=''>なし（ルート）</option>
            {categories.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </div>
        <button
          type='submit'
          className='bg-stone-900 text-white dark:bg-stone-100 dark:text-stone-900 px-5 py-2 rounded-lg text-sm font-medium hover:bg-stone-800 dark:hover:bg-stone-200 transition-colors'
        >
          作成
        </button>
      </form>

      {/* 一覧 */}
      <div className='space-y-1'>
        {categories.length === 0 ? (
          <p className='text-stone-400 text-sm text-center py-8'>カテゴリがありません</p>
        ) : (
          categories.map((c) => {
            const isRoot = !c.parent_id;
            return (
              <div
                key={c.id}
                className={`flex items-center justify-between py-3 px-4 rounded-xl transition-colors ${
                  isRoot
                    ? 'bg-stone-50 dark:bg-stone-800/50'
                    : 'bg-stone-50/50 dark:bg-stone-800/30 ml-6'
                } hover:bg-stone-100 dark:hover:bg-stone-700/50`}
              >
                <span className='flex items-center gap-2 text-sm font-medium text-stone-700 dark:text-stone-300'>
                  {isRoot ? <FiFolder size={16} /> : <FiFile size={16} />}
                  {c.name}
                </span>
                <button
                  onClick={() => setDeleteTarget(c)}
                  className='text-stone-400 hover:text-stone-600 dark:hover:text-stone-300 p-1 rounded hover:bg-stone-100 dark:hover:bg-stone-800 transition-colors'
                  title='削除'
                >
                  <FiTrash2 size={14} />
                </button>
              </div>
            );
          })
        )}
      </div>

      {deleteTarget && (
        <ConfirmModal
          isOpen={true}
          title='カテゴリ削除'
          message={`「${deleteTarget.name}」を削除しますか？`}
          onConfirm={async () => {
            await handleDelete(deleteTarget.id);
          }}
          onCancel={() => setDeleteTarget(null)}
        />
      )}
    </div>
  );
}
