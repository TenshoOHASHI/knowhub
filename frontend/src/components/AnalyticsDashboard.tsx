'use client';

import { useState, useEffect } from 'react';
import {
  FiEye,
  FiTrendingUp,
  FiFileText,
  FiLoader,
  FiUsers,
  FiHeart,
} from 'react-icons/fi';
import Link from 'next/link';

const API_BASE =
  typeof window === 'undefined'
    ? process.env.SERVER_API_URL || 'http://localhost:8080/api'
    : '/api';

interface DailyView {
  date: string;
  count: number;
  uniqueVisitors: number;
}

interface PageRankItem {
  path: string;
  count: number;
}

interface ArticleRankItem {
  id: string;
  title: string;
  count: number;
  visibility?: string;
}

interface LikeRankItem {
  id: string;
  title: string;
  count: number;
}

interface Summary {
  totalViews: number;
  uniqueVisitors: number;
  todayViews: number;
  dailyViews: DailyView[];
  pageRanking: PageRankItem[];
  articleRanking: ArticleRankItem[];
  likeRanking: LikeRankItem[];
}

export default function AnalyticsDashboard() {
  const [summary, setSummary] = useState<Summary | null>(null);
  const [days, setDays] = useState(30);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      setLoading(true);
      setError(null);
      try {
        const res = await fetch(`${API_BASE}/analytics/summary?days=${days}`);
        if (res.ok) {
          const data = await res.json();
          setSummary(data);
        } else {
          setError(`API Error: ${res.status}`);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [days]);

  if (!summary && !loading) {
    return (
      <div className='text-center py-12 text-stone-400'>
        <FiEye className='mx-auto mb-2 opacity-50' size={32} />
        <p>アナリティクスデータを取得できませんでした</p>
        {error && <p className='text-xs mt-2 text-red-400'>エラー: {error}</p>}
      </div>
    );
  }

  const maxDaily = summary
    ? Math.max(...(summary.dailyViews?.map((d) => d.count) || [1]), 1)
    : 1;

  return (
    <div className='max-w-7xl mx-auto space-y-6'>
      {/* Period selector */}
      <div className='flex items-center gap-3'>
        <div className='flex gap-1'>
          {[7, 30, 90].map((d) => (
            <button
              key={d}
              onClick={() => {
                if (!loading) {
                  setDays(d);
                }
              }}
              disabled={loading}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                days === d
                  ? 'bg-stone-900 text-white dark:bg-stone-100 dark:text-stone-900'
                  : 'text-stone-600 dark:text-stone-400 hover:bg-stone-100 dark:hover:bg-stone-800 disabled:opacity-50'
              }`}
            >
              過去{d}日間
            </button>
          ))}
        </div>
        {loading && (
          <div className='flex items-center gap-2 text-stone-400 text-sm'>
            <FiLoader className='animate-spin' size={14} />
            <span>読み込み中...</span>
          </div>
        )}
      </div>

      {/* Summary cards */}
      <div className='grid grid-cols-1 sm:grid-cols-3 gap-4'>
        <div className='p-5 rounded-lg border border-stone-200 dark:border-stone-700 bg-stone-50 dark:bg-stone-800/50'>
          <div className='flex items-center gap-2 text-stone-500 dark:text-stone-400 text-xs font-medium mb-2'>
            <FiEye size={14} />
            総ページビュー
          </div>
          <p className='text-3xl font-bold text-stone-900 dark:text-stone-100'>
            {summary ? (
              (summary.totalViews ?? 0).toLocaleString()
            ) : (
              <span className='inline-block w-16 h-8 bg-stone-200 dark:bg-stone-700 animate-pulse rounded' />
            )}
          </p>
          <p className='text-xs text-stone-500 dark:text-stone-400 mt-1'>
            過去{days}日間
          </p>
        </div>
        <div className='p-5 rounded-lg border border-stone-200 dark:border-stone-700 bg-stone-50 dark:bg-stone-800/50'>
          <div className='flex items-center gap-2 text-stone-500 dark:text-stone-400 text-xs font-medium mb-2'>
            <FiUsers size={14} />
            ユニーク訪問者
          </div>
          <p className='text-3xl font-bold text-stone-900 dark:text-stone-100'>
            {summary ? (
              (summary.uniqueVisitors ?? 0).toLocaleString()
            ) : (
              <span className='inline-block w-16 h-8 bg-stone-200 dark:bg-stone-700 animate-pulse rounded' />
            )}
          </p>
          <p className='text-xs text-stone-500 dark:text-stone-400 mt-1'>
            ユニークIP
          </p>
        </div>
        <div className='p-5 rounded-lg border border-stone-200 dark:border-stone-700 bg-stone-50 dark:bg-stone-800/50'>
          <div className='flex items-center gap-2 text-stone-500 dark:text-stone-400 text-xs font-medium mb-2'>
            <FiTrendingUp size={14} />
            今日のページビュー
          </div>
          <p className='text-3xl font-bold text-stone-900 dark:text-stone-100'>
            {summary ? (
              (summary.todayViews ?? 0).toLocaleString()
            ) : (
              <span className='inline-block w-16 h-8 bg-stone-200 dark:bg-stone-700 animate-pulse rounded' />
            )}
          </p>
          <p className='text-xs text-stone-500 dark:text-stone-400 mt-1'>
            本日
          </p>
        </div>
      </div>

      {/* Daily chart */}
      <div className='p-5 rounded-lg border border-stone-200 dark:border-stone-700 bg-stone-50 dark:bg-stone-800/50'>
        <h3 className='text-sm font-semibold text-stone-700 dark:text-stone-300 mb-4'>
          日別推移
        </h3>
        {summary && summary.dailyViews && summary.dailyViews.length > 0 ? (
          <div>
            <div
              className='flex items-end justify-between gap-1'
              style={{ height: '120px' }}
            >
              {summary.dailyViews.map((d, i) => {
                return (
                  <div key={i} className='flex-1 flex flex-col items-center'>
                    <div className='text-xs text-stone-700 dark:text-stone-300 mb-1'>
                      {d.count}
                    </div>
                    <div
                      className='w-full bg-stone-200 dark:bg-stone-800 rounded-t'
                      style={{ height: '100px' }}
                    >
                      <div
                        className='w-full bg-stone-700 dark:bg-stone-500 rounded-t hover:bg-stone-800 dark:hover:bg-stone-400 transition-colors'
                        style={{
                          height: `${(d.count / maxDaily) * 100}%`,
                          minHeight: '2px',
                        }}
                      />
                    </div>
                  </div>
                );
              })}
            </div>
            <div className='flex justify-between gap-1 mt-1'>
              {summary.dailyViews.map((d, i) => (
                <div
                  key={i}
                  className='flex-1 text-center text-[10px] text-stone-400'
                >
                  {d.date.slice(5).replace('-', '/')}
                </div>
              ))}
            </div>
          </div>
        ) : (
          <p className='text-stone-400 text-sm text-center py-8'>
            データがありません
          </p>
        )}
      </div>

      {/* Article ranking */}
      <div className='p-5 rounded-lg border border-stone-200 dark:border-stone-700 bg-stone-50 dark:bg-stone-800/50'>
        <h3 className='flex items-center gap-2 text-sm font-semibold text-stone-700 dark:text-stone-300 mb-4'>
          <FiFileText size={16} />
          人気記事 Top 10
        </h3>
        {summary &&
        summary.articleRanking &&
        summary.articleRanking.length > 0 ? (
          <div className='space-y-2'>
            {summary.articleRanking.map((a, i) => (
              <div key={i} className='flex items-center gap-3 group'>
                <span
                  className={`flex-shrink-0 w-6 h-6 flex items-center justify-center text-xs font-bold rounded-full ${
                    i === 0
                      ? 'bg-stone-800 text-white dark:bg-stone-200 dark:text-stone-900'
                      : i === 1
                        ? 'bg-stone-600 text-white dark:bg-stone-300 dark:text-stone-900'
                        : i === 2
                          ? 'bg-stone-500 text-white dark:bg-stone-400 dark:text-stone-900'
                          : 'bg-stone-200 dark:bg-stone-700 text-stone-600 dark:text-stone-400'
                  }`}
                >
                  {i + 1}
                </span>
                <div className='flex-1 min-w-0'>
                  <div className='flex items-center gap-2'>
                    <Link
                      href={`/wiki/${a.id}`}
                      className='text-sm text-stone-700 dark:text-stone-300 hover:text-stone-900 dark:hover:text-stone-100 block truncate'
                    >
                      {a.title}
                    </Link>
                    <span className='text-[10px] px-1.5 py-0.5 rounded bg-stone-200 dark:bg-stone-700 text-stone-500 dark:text-stone-400'>
                      {a.visibility || '?'}
                    </span>
                  </div>
                  <div className='h-1.5 bg-stone-200 dark:bg-stone-700 rounded-full overflow-hidden mt-1'>
                    <div
                      className='h-full bg-stone-600 dark:bg-stone-400 rounded-full transition-all duration-500'
                      style={{
                        width: `${(a.count / (summary.articleRanking[0]?.count || 1)) * 100}%`,
                      }}
                    />
                  </div>
                </div>
                <span className='text-sm font-semibold text-stone-600 dark:text-stone-400'>
                  {a.count.toLocaleString()}
                </span>
              </div>
            ))}
          </div>
        ) : (
          <p className='text-stone-400 text-sm text-center py-8'>
            公開記事のデータがありません (記事数:{' '}
            {summary?.articleRanking?.length || 0})
          </p>
        )}
      </div>

      {/* Like ranking */}
      <div className='p-5 rounded-lg border border-stone-200 dark:border-stone-700 bg-stone-50 dark:bg-stone-800/50'>
        <h3 className='flex items-center gap-2 text-sm font-semibold text-stone-700 dark:text-stone-300 mb-4'>
          <FiHeart size={16} />
          いいね数 Top 10
        </h3>
        {summary && summary.likeRanking && summary.likeRanking.length > 0 ? (
          <div className='space-y-2'>
            {summary.likeRanking.map((a, i) => (
              <div key={i} className='flex items-center gap-3 group'>
                <span
                  className={`flex-shrink-0 w-6 h-6 flex items-center justify-center text-xs font-bold rounded-full ${
                    i === 0
                      ? 'bg-rose-600 text-white dark:bg-rose-400 dark:text-stone-900'
                      : i === 1
                        ? 'bg-rose-500 text-white dark:bg-rose-300 dark:text-stone-900'
                        : i === 2
                          ? 'bg-rose-400 text-white dark:bg-rose-200 dark:text-stone-900'
                          : 'bg-rose-100 dark:bg-rose-900/50 text-rose-700 dark:text-rose-300'
                  }`}
                >
                  {i + 1}
                </span>
                <div className='flex-1 min-w-0'>
                  <Link
                    href={`/wiki/${a.id}`}
                    className='text-sm text-stone-700 dark:text-stone-300 hover:text-stone-900 dark:hover:text-stone-100 block truncate'
                  >
                    {a.title}
                  </Link>
                  <div className='h-1.5 bg-stone-200 dark:bg-stone-700 rounded-full overflow-hidden mt-1'>
                    <div
                      className='h-full bg-rose-500 dark:bg-rose-400 rounded-full transition-all duration-500'
                      style={{
                        width: `${(a.count / (summary.likeRanking[0]?.count || 1)) * 100}%`,
                      }}
                    />
                  </div>
                </div>
                <span className='text-sm font-semibold text-stone-600 dark:text-stone-400'>
                  {a.count.toLocaleString()}
                </span>
              </div>
            ))}
          </div>
        ) : (
          <p className='text-stone-400 text-sm text-center py-8'>
            いいねのデータがありません
          </p>
        )}
      </div>

      {/* Page ranking */}
      <div className='p-5 rounded-lg border border-stone-200 dark:border-stone-700 bg-stone-50 dark:bg-stone-800/50'>
        <h3 className='flex items-center gap-2 text-sm font-semibold text-stone-700 dark:text-stone-300 mb-4'>
          <FiTrendingUp size={16} />
          人気ページ Top 10
        </h3>
        {summary && summary.pageRanking && summary.pageRanking.length > 0 ? (
          <div className='space-y-2'>
            {summary.pageRanking.map((p, i) => (
              <div key={i} className='flex items-center gap-3 group'>
                <span
                  className={`flex-shrink-0 w-6 h-6 flex items-center justify-center text-xs font-bold rounded-full ${
                    i === 0
                      ? 'bg-stone-800 text-white dark:bg-stone-200 dark:text-stone-900'
                      : i === 1
                        ? 'bg-stone-600 text-white dark:bg-stone-300 dark:text-stone-900'
                        : i === 2
                          ? 'bg-stone-500 text-white dark:bg-stone-400 dark:text-stone-900'
                          : 'bg-stone-200 dark:bg-stone-700 text-stone-600 dark:text-stone-400'
                  }`}
                >
                  {i + 1}
                </span>
                <div className='flex-1 min-w-0'>
                  <span className='text-sm text-stone-700 dark:text-stone-300 block truncate'>
                    {p.path}
                  </span>
                  <div className='h-1.5 bg-stone-200 dark:bg-stone-700 rounded-full overflow-hidden mt-1'>
                    <div
                      className='h-full bg-stone-600 dark:bg-stone-400 rounded-full transition-all duration-500'
                      style={{
                        width: `${(p.count / (summary.pageRanking[0]?.count || 1)) * 100}%`,
                      }}
                    />
                  </div>
                </div>
                <span className='text-sm font-semibold text-stone-600 dark:text-stone-400'>
                  {p.count.toLocaleString()}
                </span>
              </div>
            ))}
          </div>
        ) : (
          <p className='text-stone-400 text-sm text-center py-8'>
            データがありません
          </p>
        )}
      </div>
    </div>
  );
}
