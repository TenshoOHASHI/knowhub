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
      <div className='flex'>
        {/* エディタ + プレビュー */}
        <div className='flex-1 grid grid-cols-2 gap-4'>
          <div>
            <label className='block text-sm font-medium mb-1'>
              本文（Markdown）
            </label>
            <textarea
              name='content'
              value={content}
              onChange={(e) => onContentChange(e.target.value)}
              className='w-full border border-black dark:border-stone-600 rounded-lg p-2 h-80 bg-transparent font-mono text-sm thin-scrollbar'
              placeholder='Markdownで記事を書く...'
            />
          </div>
          <div>
            <div className='flex items-center justify-between mb-1'>
              <label className='block text-sm font-medium'>プレビュー</label>
              <div className='flex gap-1'>
                <button
                  type='button'
                  onClick={() => setShowHelp(!showHelp)}
                  className='p-1 text-gray-400 hover:text-black dark:hover:text-stone-100'
                  title='Markdownリファレンス'
                >
                  <FiHelpCircle size={16} />
                </button>
                <button
                  type='button'
                  onClick={() => setShowFullPreview(true)}
                  className='p-1 text-gray-400 hover:text-black dark:hover:text-stone-100'
                  title='拡大表示'
                >
                  <FiMaximize2 size={16} />
                </button>
              </div>
            </div>
            <div className='border border-black dark:border-stone-600 rounded-lg p-2 h-80 overflow-y-auto prose dark:prose-invert thin-scrollbar'>
              {processed ? (
                <ReactMarkdown
                  remarkPlugins={[remarkGfm, remarkCallout]}
                  rehypePlugins={[rehypeRaw, rehypeHighlight]}
                  components={markdownComponents()}
                >
                  {processed}
                </ReactMarkdown>
              ) : (
                <p className='text-gray-400'>
                  入力するとここにプレビューが表示されます
                </p>
              )}
            </div>
          </div>
        </div>

        {/* ヘルプパネル */}
        {showHelp && (
          <div className='w-64 shrink-0 h-84'>
            <MarkdownHelp
              isOpen={showHelp}
              onClose={() => setShowHelp(false)}
            />
          </div>
        )}
      </div>

      {/* 拡大プレビューモーダル */}
      {showFullPreview && (
        <div
          className='fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-8'
          onClick={() => setShowFullPreview(false)}
        >
          <div
            className='bg-white dark:bg-stone-900 rounded-lg w-full max-w-4xl max-h-[85vh] overflow-y-auto p-8 shadow-xl thin-scrollbar'
            onClick={(e) => e.stopPropagation()}
          >
            <div className='flex justify-between items-center mb-4'>
              <h2 className='text-xl font-bold'>プレビュー</h2>
              <button
                onClick={() => setShowFullPreview(false)}
                className='text-gray-400 hover:text-black dark:hover:text-stone-100'
              >
                <FiX size={20} />
              </button>
            </div>
            <div className='prose max-w-none dark:prose-invert'>
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
    </>
  );
}
