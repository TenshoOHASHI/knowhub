'use client';

import { useState, useEffect, useCallback } from 'react';
import { useToast } from '@/context/ToastContext';
import type { Category } from '@/lib/types';
import { getCategories, createCategory, deleteCategory } from '@/lib/api';
import ConfirmModal from './ConfirmModal';
import { FiFile, FiFolder, FiTrash2 } from 'react-icons/fi';

/** カテゴリの深さを算出（ルート=0, 子=1, 孫=2） */
function getDepth(categoryId: string, categories: Category[]): number {
  const map = new Map(categories.map((c) => [c.id, c]));
  let depth = 0;
  let current = map.get(categoryId);
  while (current?.parent_id) {
    depth++;
    current = map.get(current.parent_id);
    if (depth > 10) break; // 無限ループ防止
  }
  return depth;
}

/** フラットなカテゴリ配列を「親 → 子 → 孫…」の順に3階層まで展開 */
function sortByTree(categories: Category[]): Category[] {
  const roots: Category[] = [];
  const childrenMap = new Map<string, Category[]>();

  for (const c of categories) {
    if (!c.parent_id) {
      roots.push(c);
    } else {
      const list = childrenMap.get(c.parent_id) || [];
      list.push(c);
      childrenMap.set(c.parent_id, list);
    }
  }

  const result: Category[] = [];
  for (const root of roots) {
    result.push(root);
    const children = childrenMap.get(root.id) || [];
    for (const child of children) {
      result.push(child);
      const grandchildren = childrenMap.get(child.id) || [];
      result.push(...grandchildren);
    }
  }

  // 親が存在しない孤立した子カテゴリも末尾に追加
  const added = new Set(result.map((c) => c.id));
  for (const c of categories) {
    if (!added.has(c.id)) {
      result.push(c);
    }
  }

  return result;
}

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
            {categories.map((c) => {
              const depth = getDepth(c.id, categories);
              return (
                <option key={c.id} value={c.id} disabled={depth >= 2}>
                  {'　'.repeat(depth)}{c.name}{depth >= 2 ? '（これ以上の階層は作成不可）' : ''}
                </option>
              );
            })}
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
          sortByTree(categories).map((c) => {
            const depth = getDepth(c.id, categories);
            return (
              <div
                key={c.id}
                className={`flex items-center justify-between py-3 px-4 rounded-xl transition-colors ${
                  depth === 0
                    ? 'bg-stone-50 dark:bg-stone-800/50'
                    : depth === 1
                      ? 'bg-stone-50/50 dark:bg-stone-800/30 ml-6'
                      : 'bg-stone-50/30 dark:bg-stone-800/20 ml-12'
                } hover:bg-stone-100 dark:hover:bg-stone-700/50`}
              >
                <span className='flex items-center gap-2 text-sm font-medium text-stone-700 dark:text-stone-300'>
                  {depth === 0 ? <FiFolder size={16} /> : <FiFile size={16} />}
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
