'use client';

import { Suspense, useEffect, useState } from 'react';

import Editor from '@/components/Editor';
import { CategoryManager } from '@/components/CategoryManager';
import ProfileManager from '@/components/ProfileManager';
import { PortfolioManager } from '@/components/PortfolioManager';
import AnalyticsDashboard from '@/components/AnalyticsDashboard';
import LogViewer from '@/components/LogViewer';
import { useAuth } from '@/context/AuthContext';
import { useRouter } from 'next/navigation';
import { FiFileText, FiArchive, FiUser, FiBriefcase, FiTrendingUp, FiTerminal } from 'react-icons/fi';

const tabs = [
  { id: 'article' as const, label: '記事作成', icon: FiFileText, bg: 'bg-stone-100 dark:bg-stone-800', text: 'text-stone-600 dark:text-stone-400' },
  { id: 'category' as const, label: 'カテゴリ', icon: FiArchive, bg: 'bg-stone-100 dark:bg-stone-800', text: 'text-stone-600 dark:text-stone-400' },
  { id: 'profile' as const, label: 'プロフィール', icon: FiUser, bg: 'bg-stone-100 dark:bg-stone-800', text: 'text-stone-600 dark:text-stone-400' },
  { id: 'portfolio' as const, label: 'ポートフォリオ', icon: FiBriefcase, bg: 'bg-stone-100 dark:bg-stone-800', text: 'text-stone-600 dark:text-stone-400' },
  { id: 'analytics' as const, label: 'アナリティクス', icon: FiTrendingUp, bg: 'bg-stone-100 dark:bg-stone-800', text: 'text-stone-600 dark:text-stone-400' },
  { id: 'logs' as const, label: 'ログ', icon: FiTerminal, bg: 'bg-stone-100 dark:bg-stone-800', text: 'text-stone-600 dark:text-stone-400' },
];

export default function AdminPage() {
  const [activeTab, setActiveTab] = useState<
    'article' | 'category' | 'profile' | 'portfolio' | 'analytics' | 'logs'
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

  const currentTab = tabs.find(t => t.id === activeTab)!;

  return (
    <div className='max-w-7xl mx-auto p-6'>
      {/* ヘッダー */}
      <div className='mb-6'>
        <div className='flex items-center gap-4 mb-6 pb-4 border-b border-stone-200 dark:border-stone-700'>
          <div className={`w-11 h-11 rounded-lg ${currentTab.bg} flex items-center justify-center`}>
            <currentTab.icon size={20} className={currentTab.text} />
          </div>
          <div>
            <h1 className='text-xl font-semibold text-stone-900 dark:text-stone-100'>
              管理画面
            </h1>
            <p className='text-xs text-stone-500 dark:text-stone-400 mt-0.5'>
              {currentTab.label}の管理・編集
            </p>
          </div>
        </div>

        {/* タブ切り替え */}
        <div className='flex flex-wrap gap-1'>
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                activeTab === tab.id
                  ? 'bg-stone-900 text-white dark:bg-stone-100 dark:text-stone-900'
                  : 'text-stone-600 hover:bg-stone-100 dark:text-stone-400 dark:hover:bg-stone-800'
              }`}
            >
              <tab.icon size={15} />
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* コンテンツエリア */}
      {activeTab === 'article' ? (
        <Suspense fallback={null}>
          <Editor embedded={true} />
        </Suspense>
      ) : activeTab === 'category' ? (
        <div className='rounded-xl border border-stone-200 dark:border-stone-700 bg-white dark:bg-stone-800 shadow-sm overflow-hidden'>
          <CategoryManager />
        </div>
      ) : activeTab === 'profile' ? (
        <div className='rounded-xl border border-stone-200 dark:border-stone-700 bg-white dark:bg-stone-800 shadow-sm overflow-hidden'>
          <ProfileManager />
        </div>
      ) : activeTab === 'analytics' ? (
        <AnalyticsDashboard />
      ) : activeTab === 'logs' ? (
        <LogViewer />
      ) : (
        <div className='rounded-xl border border-stone-200 dark:border-stone-700 bg-white dark:bg-stone-800 shadow-sm overflow-hidden'>
          <PortfolioManager />
        </div>
      )}
    </div>
  );
}
