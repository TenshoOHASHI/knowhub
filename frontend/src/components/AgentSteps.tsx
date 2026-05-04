'use client';

import { useState } from 'react';
import { FaChevronDown, FaChevronRight, FaSearch, FaBookOpen, FaGlobe, FaList } from 'react-icons/fa';
import ReactMarkdown from 'react-markdown';
import type { AgentStep, AgentSource } from '@/lib/api';

interface AgentStepsProps {
  steps: AgentStep[];
  sources: AgentSource[];
}

// ツール名の表示用マッピング
const toolMeta: Record<string, { label: string; color: string; icon: typeof FaSearch }> = {
  search_wiki: { label: 'Wiki検索', color: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300', icon: FaSearch },
  read_article: { label: '記事取得', color: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300', icon: FaBookOpen },
  list_articles: { label: '記事一覧', color: 'bg-stone-100 text-stone-700 dark:bg-stone-700 dark:text-stone-300', icon: FaList },
  web_search: { label: 'Web検索', color: 'bg-amber-100 text-amber-700 dark:bg-amber-900 dark:text-amber-300', icon: FaGlobe },
  read_url: { label: 'URL取得', color: 'bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300', icon: FaGlobe },
};

export default function AgentSteps({ steps, sources }: AgentStepsProps) {
  const [open, setOpen] = useState(false);
  const [openObs, setOpenObs] = useState<Record<number, boolean>>({});

  if (steps.length === 0 && sources.length === 0) return null;

  return (
    <div className='mt-2'>
      <button
        type='button'
        onClick={() => setOpen(!open)}
        className='flex items-center gap-1 text-xs text-stone-500 dark:text-stone-400 hover:text-stone-700 dark:hover:text-stone-200 transition-colors'
      >
        {open ? <FaChevronDown size={10} /> : <FaChevronRight size={10} />}
        思考プロセス ({steps.length} ステップ)
      </button>

      {open && (
        <div className='mt-2 space-y-3 text-xs'>
          {steps.map((step, i) => {
            const meta = toolMeta[step.action];
            const Icon = meta?.icon || FaSearch;
            const isObsLong = step.observation && step.observation.length > 200;
            const isObsOpen = openObs[i] ?? false;

            return (
              <div
                key={i}
                className='border border-stone-200 dark:border-stone-600 rounded-lg p-3 space-y-2'
              >
                {/* ステップヘッダー */}
                <div className='flex items-center gap-2'>
                  <span className='text-stone-400 text-[10px] font-mono'>
                    #{i + 1}
                  </span>
                  {meta ? (
                    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium ${meta.color}`}>
                      <Icon size={10} />
                      {meta.label}
                    </span>
                  ) : step.action ? (
                    <span className='px-2 py-0.5 rounded-full text-[10px] font-medium bg-stone-100 text-stone-600 dark:bg-stone-700 dark:text-stone-300'>
                      {step.action}
                    </span>
                  ) : null}
                </div>

                {/* 思考 */}
                {step.thought && (
                  <div className='text-stone-600 dark:text-stone-300'>
                    {step.thought}
                  </div>
                )}

                {/* 入力 */}
                {step.action_input && (
                  <code className='block bg-stone-100 dark:bg-stone-700 rounded px-2 py-1 text-[11px] break-all text-stone-700 dark:text-stone-300'>
                    {step.action_input}
                  </code>
                )}

                {/* 観察結果 */}
                {step.observation && (
                  <div>
                    {isObsLong && !isObsOpen ? (
                      <div className='bg-stone-50 dark:bg-stone-700/50 rounded px-2 py-1'>
                        <div className='prose prose-sm prose-p:my-1 prose-headings:my-1 prose-li:my-0.5 max-w-none text-stone-600 dark:text-stone-300 [&_h1]:text-sm [&_h2]:text-sm [&_h3]:text-xs line-clamp-4'>
                          <ReactMarkdown>{step.observation}</ReactMarkdown>
                        </div>
                        <button
                          type='button'
                          onClick={() => setOpenObs((prev) => ({ ...prev, [i]: true }))}
                          className='text-blue-500 hover:text-blue-400 text-[10px]'
                        >
                          もっと見る
                        </button>
                      </div>
                    ) : (
                      <div className='bg-stone-50 dark:bg-stone-700/50 rounded px-2 py-1'>
                        <div className='prose prose-sm prose-p:my-1 prose-headings:my-1 prose-li:my-0.5 max-w-none text-stone-600 dark:text-stone-300 [&_h1]:text-sm [&_h2]:text-sm [&_h3]:text-xs'>
                          <ReactMarkdown>{step.observation}</ReactMarkdown>
                        </div>
                        {isObsLong && (
                          <button
                            type='button'
                            onClick={() => setOpenObs((prev) => ({ ...prev, [i]: false }))}
                            className='text-blue-500 hover:text-blue-400 text-[10px]'
                          >
                            閉じる
                          </button>
                        )}
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })}

          {sources.length > 0 && (
            <div className='pt-2 border-t border-stone-200 dark:border-stone-600'>
              <span className='text-stone-400 text-[10px] font-medium uppercase tracking-wider'>
                Sources
              </span>
              <div className='flex flex-wrap gap-1.5 mt-1'>
                {sources.map((s, i) => (
                  <span key={i}>
                    {s.article_id ? (
                      <a
                        href={`/wiki/${s.article_id}`}
                        className='inline-block px-2 py-0.5 rounded-full bg-stone-100 dark:bg-stone-700 text-stone-600 dark:text-stone-300 hover:bg-blue-100 dark:hover:bg-blue-900 hover:text-blue-600 dark:hover:text-blue-300 transition-colors'
                      >
                        {s.title || s.article_id}
                      </a>
                    ) : s.url ? (
                      <a
                        href={s.url}
                        target='_blank'
                        rel='noopener noreferrer'
                        className='inline-block px-2 py-0.5 rounded-full bg-stone-100 dark:bg-stone-700 text-stone-600 dark:text-stone-300 hover:bg-amber-100 dark:hover:bg-amber-900 hover:text-amber-600 dark:hover:text-amber-300 transition-colors'
                      >
                        {(() => {
                          try {
                            return new URL(s.url).hostname;
                          } catch {
                            return s.url;
                          }
                        })()}
                      </a>
                    ) : null}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
