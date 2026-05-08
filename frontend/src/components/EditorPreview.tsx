'use client';

import { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import rehypeHighlight from 'rehype-highlight';
import rehypeRaw from 'rehype-raw';
import markdownComponents from './Markdown';
import MarkdownHelp from './MarkdownHelp';
import { FiHelpCircle, FiMaximize2, FiX } from 'react-icons/fi';
import remarkGfm from 'remark-gfm';
import { remarkCallout, preprocessCallouts } from '@/lib/remark-callout';

// 親（Editor）から渡される props の型定義
interface EditorPreviewProps {
  /** テキストエリアの現在の値（親の state と同期） */
  content: string;
  /** テキストエリアの値が変わったときに親の state を更新するコールバック */
  onContentChange: (value: string) => void;
}

export default function EditorPreview({
  content,
  onContentChange,
}: EditorPreviewProps) {
  const [showFullPreview, setShowFullPreview] = useState(false);
  const [showHelp, setShowHelp] = useState(false);

  const processed = preprocessCallouts(content);

  return (
    <>
      {/* エディタ + プレビュー */}
      <div className='rounded-xl border border-stone-200 dark:border-stone-700 bg-white dark:bg-stone-800 shadow-sm overflow-hidden'>
        <div className='grid grid-cols-2 divide-x divide-stone-200 dark:divide-stone-700'>
          {/* エディタ */}
          <div className='p-4'>
            <div className='flex items-center justify-between mb-2'>
              <label className='block text-sm font-semibold text-stone-900 dark:text-stone-100'>
                本文（Markdown）
              </label>
            </div>
            <textarea
              name='content'
              value={content}
              onChange={(e) => onContentChange(e.target.value)}
              className='w-full border border-stone-300 dark:border-stone-600 rounded-xl p-3 h-80 bg-stone-50 dark:bg-stone-900 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 thin-scrollbar resize-none'
              placeholder='Markdownで記事を書く...'
            />
          </div>

          {/* プレビュー */}
          <div className='p-4 bg-stone-50/50 dark:bg-stone-900/30'>
            <div className='flex items-center justify-between mb-2'>
              <label className='block text-sm font-semibold text-stone-900 dark:text-stone-100'>プレビュー</label>
              <div className='flex gap-1'>
                <button
                  type='button'
                  onClick={() => setShowHelp(true)}
                  className='p-1.5 text-stone-400 hover:text-stone-600 dark:hover:text-stone-200 hover:bg-stone-200 dark:hover:bg-stone-700 rounded transition-colors'
                  title='Markdownリファレンス'
                >
                  <FiHelpCircle size={16} />
                </button>
                <button
                  type='button'
                  onClick={() => setShowFullPreview(true)}
                  className='p-1.5 text-stone-400 hover:text-stone-600 dark:hover:text-stone-200 hover:bg-stone-200 dark:hover:bg-stone-700 rounded transition-colors'
                  title='拡大表示'
                >
                  <FiMaximize2 size={16} />
                </button>
              </div>
            </div>
            <div className='border border-stone-200 dark:border-stone-700 rounded-xl p-3 h-80 overflow-y-auto prose prose-sm dark:prose-invert prose-li:marker:text-stone-600 dark:prose-li:marker:text-stone-300 prose-hr:border-stone-500 dark:prose-hr:border-stone-400 prose-code:text-stone-800 dark:prose-code:text-stone-200 max-w-none thin-scrollbar bg-white dark:bg-stone-800'>
              {processed ? (
                <ReactMarkdown
                  remarkPlugins={[remarkGfm, remarkCallout]}
                  rehypePlugins={[rehypeRaw, rehypeHighlight]}
                  components={markdownComponents()}
                >
                  {processed}
                </ReactMarkdown>
              ) : (
                <p className='text-stone-400 text-sm'>
                  入力するとここにプレビューが表示されます
                </p>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* 拡大プレビューモーダル */}
      {showFullPreview && (
        <div
          className='fixed inset-0 z-50 bg-black/40 backdrop-blur-sm flex items-center justify-center p-8'
          onClick={() => setShowFullPreview(false)}
        >
          <div
            className='bg-white dark:bg-stone-900 rounded-xl w-full max-w-4xl max-h-[85vh] overflow-hidden shadow-2xl thin-scrollbar flex flex-col'
            onClick={(e) => e.stopPropagation()}
          >
            <div className='flex justify-between items-center px-6 py-4 border-b border-stone-200 dark:border-stone-700 bg-stone-50 dark:bg-stone-800/50'>
              <h2 className='text-lg font-bold text-stone-900 dark:text-stone-100'>プレビュー</h2>
              <button
                onClick={() => setShowFullPreview(false)}
                className='text-stone-400 hover:text-stone-600 dark:hover:text-stone-200 p-1 rounded hover:bg-stone-200 dark:hover:bg-stone-700 transition-colors'
              >
                <FiX size={20} />
              </button>
            </div>
            <div className='flex-1 overflow-y-auto p-6 prose max-w-none dark:prose-invert'>
              <ReactMarkdown
                remarkPlugins={[remarkGfm, remarkCallout]}
                rehypePlugins={[rehypeRaw, rehypeHighlight]}
                components={markdownComponents()}
              >
                {processed}
              </ReactMarkdown>
            </div>
          </div>
        </div>
      )}

      {/* Markdownヘルプモーダル */}
      <MarkdownHelp isOpen={showHelp} onClose={() => setShowHelp(false)} />
    </>
  );
}
