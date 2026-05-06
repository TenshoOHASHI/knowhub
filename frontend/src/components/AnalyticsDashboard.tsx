'use client';

import { useState, useEffect } from 'react';
import { FiEye, FiTrendingUp, FiFileText } from 'react-icons/fi';

const API_BASE =
  typeof window === 'undefined'
    ? process.env.SERVER_API_URL || 'http://localhost:8080/api'
    : '/api';

interface DailyView {
  date: string;
  count: number;
}

interface PageRankItem {
  path: string;
  count: number;
}

interface Summary {
  totalViews: number;
  todayViews: number;
  dailyViews: DailyView[];
  pageRanking: PageRankItem[];
}

export default function AnalyticsDashboard() {
  const [summary, setSummary] = useState<Summary | null>(null);
  const [days, setDays] = useState(30);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const res = await fetch(`${API_BASE}/analytics/summary?days=${days}`);
        if (res.ok) {
          const data = await res.json();
          setSummary(data);
        }
      } catch {
        // Silently fail
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [days]);

  if (loading) {
    return (
      <div className='space-y-4 max-w-7xl m'>
        {[1, 2, 3].map((i) => (
          <div
            key={i}
            className='h-24 bg-stone-100 dark:bg-stone-800 rounded-lg animate-pulse'
          />
        ))}
      </div>
    );
  }

  if (!summary) {
    return (
      <div className='text-center py-8 text-stone-400'>
        アナリティクスデータを取得できませんでした
      </div>
    );
  }

  const maxDaily = Math.max(
    ...(summary.dailyViews?.map((d) => d.count) || [1]),
    1,
  );

  return (
    <div className='max-w-7xl mx-auto space-y-6'>
      {/* Period selector */}
      <div className='flex gap-2'>
        {[7, 30, 90].map((d) => (
          <button
            key={d}
            onClick={() => {
              setDays(d);
              setLoading(true);
            }}
            className={`px-3 py-1 rounded text-sm ${
              days === d
                ? 'bg-stone-900 text-white dark:bg-stone-100 dark:text-stone-900'
                : 'text-stone-500 hover:bg-stone-100 dark:hover:bg-stone-800'
            }`}
          >
            過去{d}日
          </button>
        ))}
      </div>

      {/* Summary cards */}
      <div className='grid grid-cols-2 gap-4'>
        <div className='p-4 rounded-lg border border-stone-200 dark:border-stone-700'>
          <div className='flex items-center gap-2 text-stone-400 text-sm mb-1'>
            <FiEye size={14} />
            総訪問数
          </div>
          <p className='text-2xl font-bold'>
            {(summary.totalViews ?? 0).toLocaleString()}
          </p>
        </div>
        <div className='p-4 rounded-lg border border-stone-200 dark:border-stone-700'>
          <div className='flex items-center gap-2 text-stone-400 text-sm mb-1'>
            <FiTrendingUp size={14} />
            今日の訪問数
          </div>
          <p className='text-2xl font-bold'>
            {(summary.todayViews ?? 0).toLocaleString()}
          </p>
        </div>
      </div>

      {/* Daily chart (simple bar chart) */}
      {summary.dailyViews && summary.dailyViews.length > 0 && (
        <div className='p-4 rounded-lg border border-stone-200 dark:border-stone-700'>
          <h3 className='text-sm text-stone-400 mb-3'>日別訪問数</h3>
          <div className='flex items-end gap-1 h-32'>
            {summary.dailyViews.map((d, i) => (
              <div
                key={i}
                className='flex-1 bg-blue-500/70 rounded-t hover:bg-blue-500 transition-colors relative group min-w-0'
                style={{ height: `${(d.count / maxDaily) * 100}%` }}
                title={`${d.date}: ${d.count}`}
              >
                <div className='absolute -top-6 left-1/2 -translate-x-1/2 text-xs text-stone-400 opacity-0 group-hover:opacity-100 whitespace-nowrap'>
                  {d.count}
                </div>
              </div>
            ))}
          </div>
          <div className='flex justify-between mt-1'>
            {summary.dailyViews.length > 0 && (
              <span className='text-xs text-stone-400'>
                {summary.dailyViews[0].date}
              </span>
            )}
            {summary.dailyViews.length > 1 && (
              <span className='text-xs text-stone-400'>
                {summary.dailyViews[summary.dailyViews.length - 1].date}
              </span>
            )}
          </div>
        </div>
      )}

      {/* Page ranking */}
      {summary.pageRanking && summary.pageRanking.length > 0 && (
        <div className='p-4 rounded-lg border border-stone-200 dark:border-stone-700'>
          <h3 className='flex items-center gap-2 text-sm text-stone-400 mb-3'>
            <FiFileText size={14} />
            人気ページ Top 10
          </h3>
          <div className='space-y-2'>
            {summary.pageRanking.map((p, i) => (
              <div key={i} className='flex items-center gap-3'>
                <span className='text-sm text-stone-400 w-6 text-right'>
                  {i + 1}
                </span>
                <div className='flex-1'>
                  <div className='flex items-center justify-between'>
                    <span className='text-sm font-mono truncate'>{p.path}</span>
                    <span className='text-sm text-stone-400 ml-2'>
                      {p.count}
                    </span>
                  </div>
                  <div className='h-1.5 bg-stone-100 dark:bg-stone-800 rounded-full mt-1'>
                    <div
                      className='h-full bg-blue-500/60 rounded-full'
                      style={{
                        width: `${(p.count / (summary.pageRanking[0]?.count || 1)) * 100}%`,
                      }}
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
