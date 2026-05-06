'use client';

import { Suspense, useEffect, useState } from 'react';

import Editor from '@/components/Editor';
import { CategoryManager } from '@/components/CategoryManager';
import ProfileManager from '@/components/ProfileManager';
import { PortfolioManager } from '@/components/PortfolioManager';
import AnalyticsDashboard from '@/components/AnalyticsDashboard';
import { useAuth } from '@/context/AuthContext';
import { useRouter } from 'next/navigation';

export default function AdminPage() {
  const [activeTab, setActiveTab] = useState<
    'article' | 'category' | 'profile' | 'portfolio' | 'analytics'
  >('article');
  const { isLoggedIn } = useAuth();
  const router = useRouter();

  //　ログインの状態を監視
  useEffect(() => {
    if (isLoggedIn === false) {
      router.push('/login');
    }
  }, [isLoggedIn, router]);

  if (isLoggedIn === null) return <div>Loading...</div>;

  if (!isLoggedIn) return null;

  return (
    <>
      {/* タブ切り替えボタン */}
      <div className='flex gap-2 mb-4'>
        <button
          onClick={() => setActiveTab('article')}
          className={`px-3 py-1 rounded ${activeTab === 'article' ? 'bg-stone-900 text-white dark:bg-stone-100 dark:text-stone-900' : 'text-gray-500'}`}
        >
          記事作成
        </button>
        <button
          onClick={() => setActiveTab('category')}
          className={`px-3 py-1 rounded ${activeTab === 'category' ? 'bg-stone-900 text-white dark:bg-stone-100 dark:text-stone-900' : 'text-gray-500'}`}
        >
          カテゴリ管理
        </button>
        <button
          onClick={() => setActiveTab('profile')}
          className={`px-3 py-1 rounded ${activeTab === 'profile' ? 'bg-stone-900 text-white dark:bg-stone-100 dark:text-stone-900' : 'text-gray-500'}`}
        >
          プロファイル
        </button>
        <button
          onClick={() => setActiveTab('portfolio')}
          className={`px-3 py-1 rounded ${activeTab === 'portfolio' ? 'bg-stone-900 text-white dark:bg-stone-100 dark:text-stone-900' : 'text-gray-500'}`}
        >
          ポートフォリオ
        </button>
        <button
          onClick={() => setActiveTab('analytics')}
          className={`px-3 py-1 rounded ${activeTab === 'analytics' ? 'bg-stone-900 text-white dark:bg-stone-100 dark:text-stone-900' : 'text-gray-500'}`}
        >
          アナリティクス
        </button>
      </div>

      {activeTab === 'article' ? (
        <Suspense fallback={null}>
          <Editor />
        </Suspense>
      ) : activeTab === 'category' ? (
        <CategoryManager />
      ) : activeTab === 'profile' ? (
        <ProfileManager />
      ) : activeTab === 'analytics' ? (
        <AnalyticsDashboard />
      ) : (
        <PortfolioManager />
      )}
    </>
  );
}
