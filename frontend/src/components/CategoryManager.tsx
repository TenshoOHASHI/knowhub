'use client';

import { useState, useEffect, useCallback } from 'react';
import { useToast } from '@/context/ToastContext';
import type { Category } from '@/lib/types';
import { getCategories, createCategory, deleteCategory } from '@/lib/api';
import ConfirmModal from './ConfirmModal';
import { FiFile, FiFolder } from 'react-icons/fi';

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
      load(); // 一覧を再取得
    } catch {
      toast('作成に失敗しました', 'error');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteCategory(id);
      toast('カテゴリを削除しました', 'success');
      load();
    } catch {
      toast('削除に失敗しました', 'error');
    }
  };

  return (
    <div className='space-y-4 max-w-3xl mx-auto'>
      {/* 作成フォーム */}
      <form onSubmit={handleCreate} className='flex gap-2 items-end'>
        <div>
          <label className='block text-lg font-medium mb-1'>カテゴリ名</label>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            className='border border-black dark:border-stone-600 rounded-lg h-10 bg-transparent p-2'
            placeholder='例: Go'
            required
          />
        </div>
        <div>
          <label className='block text-lg font-medium mb-1'>親カテゴリ</label>
          <select
            //　選択状態、stateから同期される
            value={parentId}
            // ユーザーが選択、stateを更新
            onChange={(e) => setParentId(e.target.value)}
            className='border border-black dark:border-stone-600 rounded-lg h-10 bg-transparent p-2'
          >
            <option value=''>なし（ルート）</option>
            {categories.map((c) => (
              // NOTE: 選択したカテゴリのidをvalueに保存、e.target.valueで使用される
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </div>
        <button
          type='submit'
          className='bg-stone-900 dark:bg-stone-100 text-white dark:text-stone-900 px-4 rounded-lg h-10 ml-2'
        >
          作成
        </button>
      </form>

      {/* 一覧 + 削除 */}
      <ul className='space-y-1'>
        {categories.map((c) => {
          const isRoot = !c.parent_id; // parent_id が空 = ルート
          return (
            <li
              key={c.id}
              className={`flex items-center justify-between py-1 px-2 rounded
          hover:bg-gray-100 dark:hover:bg-stone-800
          ${!isRoot ? 'ml-6' : ''}`} // ← 子ならインデント
            >
              <span className='flex items-center gap-2 text-lg'>
                {isRoot ? <FiFolder /> : <FiFile />}{' '}
                {/* ← ルート: フォルダ、子: ファイル */}
                {c.name}
              </span>
              <button
                onClick={() => setDeleteTarget(c)}
                className='text-red-500 text-md hover:underline'
              >
                削除
              </button>
            </li>
          );
        })}
      </ul>
      {deleteTarget && (
        <ConfirmModal
          isOpen={true}
          title='カテゴリ削除'
          message='この記事を削除しますか？'
          onConfirm={async () => {
            await handleDelete(deleteTarget.id);
            setDeleteTarget(null);
          }}
          onCancel={() => setDeleteTarget(null)}
        />
      )}
    </div>
  );
}
