'use client';

import { useSidebar } from '@/context/SidebarContext';
import { getCategories } from '@/lib/api';
import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { FiFolder } from 'react-icons/fi';
import { GoBook } from 'react-icons/go';
import type { Category } from '@/lib/types';

interface CategoryNode extends Category {
  children: CategoryNode[];
}

// ツリー構造の閉鎖管理
function CategoryTree({ nodes }: { nodes: CategoryNode[] }) {
  return (
    <ul className='space-y-1'>
      {nodes.map((node) => (
        // ノードを取り出す
        <CategoryItem key={node.id} node={node} />
      ))}
    </ul>
  );
}

// 再帰的に繰り返す
function CategoryItem({ node }: { node: CategoryNode }) {
  const [open, setOpen] = useState(false);
  const hasChildren = node.children.length > 0;
  const router = useRouter();
  const searchParams = useSearchParams();
  const selectedId = searchParams.get('category');

  const handleClick = () => {
    if (hasChildren) setOpen(!open);
    // カテゴリで絞り込み（URLパラメータを変更）
    router.push(`/wiki?category=${node.id}`);
  };

  const isSelected = selectedId === node.id;

  return (
    <li>
      {/* 子ノードが存在するなら、開く、また閉じる */}
      <button
        //  子ノードをクリックし、URLパラメータを変更し、カテゴリidを取得
        onClick={handleClick}
        className={`flex items-center gap-2 w-full text-left text-sm
            py-1 px-2 rounded
            hover:bg-gray-100 dark:hover:bg-stone-800
            ${
              isSelected
                ? 'font-bold text-black dark:text-stone-100 bg-gray-100 dark:bg-stone-800'
                : 'text-gray-600 dark:text-stone-400 hover:text-black dark:hover:text-stone-100'
            }`}
      >
        {/* 子ノードが存在し、且つ開いていれば、 */}
        {hasChildren && (open ? '▾' : '▸')}
        <FiFolder className='shrink-0' />
        <span>{node.name}</span>
      </button>
      {open && hasChildren && (
        <ul className='ml-4 mt-1 space-y-1'>
          {node.children.map((child) => (
            <CategoryItem key={child.id} node={child} />
          ))}
        </ul>
      )}
    </li>
  );
}

// 配列をマップにIDとして変換
function buildTree(categories: Category[]): CategoryNode[] {
  if (!categories || !Array.isArray(categories)) return [];
  // id → CategoryNode のマップを作る
  const map = new Map<string, CategoryNode>();
  categories.forEach((c) => map.set(c.id, { ...c, children: [] }));

  // ツリー構築
  const roots: CategoryNode[] = [];
  categories.forEach((c) => {
    const node = map.get(c.id)!;
    if (!c.parent_id) {
      roots.push(node); // ルートカテゴリ
    } else {
      const parent = map.get(c.parent_id);
      parent?.children.push(node); // 親の子として追加
    }
  });

  return roots;
}

export default function Sidebar() {
  const { isOpen } = useSidebar();
  const [categories, setCategories] = useState<Category[]>([]);
  const router = useRouter();
  const searchParams = useSearchParams();
  const selectedId = searchParams.get('category');

  useEffect(() => {
    getCategories()
      .then((data) => setCategories(data.categories || []))
      .catch(console.error);
  }, []);

  if (!isOpen) return null;

  const tree = buildTree(categories);

  return (
    <aside className='w-48 border-r border-black dark:border-stone-600 shrink-0 overflow-y-auto h-full thin-scrollbar'>
      <h2 className='font-semibold mb-3 text-md sticky top-0 bg-white dark:bg-stone-900/90 py-4 px-4 z-10'>
        カテゴリ
      </h2>
      <div className='px-4 pb-2'>
        <button
          onClick={() => router.push('/wiki')}
          className={`text-sm px-2 py-1 rounded w-full text-left
            ${
              !selectedId
                ? 'font-bold text-black dark:text-stone-100 bg-gray-100 dark:bg-stone-800'
                : 'text-gray-500 hover:text-black dark:hover:text-stone-100'
            }`}
        >
          <GoBook className='inline' /> <span className='ml-1'>すべて</span>
        </button>
      </div>
      {categories.length == 0 && (
        <p className='text-sm text-center'>データが存在しません。</p>
      )}
      <div className='px-4 pb-4'>
        <CategoryTree nodes={tree} />
      </div>
    </aside>
  );
}
