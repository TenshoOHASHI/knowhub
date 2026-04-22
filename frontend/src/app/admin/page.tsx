'use client';

import { useToast } from '@/context/ToastContext';
import { useActionState, useEffect, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import rehypeHighlight from 'rehype-highlight';
import MermaidDiagram from '@/components/MermaidDiagram';
import MarkdownHelp from '@/components/MarkdownHelp';
import { FiMaximize2, FiHelpCircle, FiX } from 'react-icons/fi';
import 'highlight.js/styles/github.css';
import 'highlight.js/styles/github-dark.css';
import remarkGfm from 'remark-gfm';

async function createArticle(prevState: string | null, formData: FormData) {
  const title = formData.get('title') as string;
  const content = formData.get('content') as string;
  const res = await fetch('http://localhost:8080/api/articles', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title, content }),
  });

  if (res.ok) {
    return '記事を作成しました';
  }
  return '作成に失敗しました';
}

// Markdownレンダリングの共通components設定
function markdownComponents() {
  return {
    code({
      className,
      children,
      ...props
    }: React.ComponentProps<'code'> & { className?: string }) {
      const match = /language-(\w+)/.exec(className || '');
      const lang = match ? match[1] : '';
      const codeString = String(children).replace(/\n$/, '');

      if (lang === 'mermaid') {
        return <MermaidDiagram chart={codeString} />;
      }

      return (
        <code className={className} {...props}>
          {children}
        </code>
      );
    },
  };
}

export default function AdminPage() {
  const [message, formAction] = useActionState(createArticle, null);
  const [preview, setPreview] = useState('');
  const [showFullPreview, setShowFullPreview] = useState(false);
  const [showHelp, setShowHelp] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    if (!message) return;
    if (message === '記事を作成しました') {
      toast(message, 'success');
    } else {
      toast(message, 'error');
    }
  }, [message, toast]);

  return (
    <>
      <div className='max-w-7xl mx-auto p-6'>
        <h1 className='text-3xl font-bold mb-4'>記事エディタ</h1>
        <form action={formAction} className='space-y-4'>
          <div>
            <label className='block text-sm font-medium mb-1'>タイトル</label>
            <input
              name='title'
              type='text'
              className='w-full border border-black dark:border-stone-600 rounded-lg p-2 bg-transparent'
              placeholder='記事タイトル'
            />
          </div>

          <div className='flex'>
            {/* エディタ + プレビュー */}
            <div className='flex-1 grid grid-cols-2 gap-4'>
              <div>
                <label className='block text-sm font-medium mb-1'>
                  本文（Markdown）
                </label>
                <textarea
                  name='content'
                  onChange={(e) => setPreview(e.target.value)}
                  className='w-full border border-black dark:border-stone-600 rounded-lg p-2 h-80 bg-transparent font-mono text-sm thin-scrollbar'
                  placeholder='Markdownで記事を書く...'
                />
              </div>
              <div>
                <div className='flex items-center justify-between mb-1'>
                  <label className='block text-sm font-medium'>
                    プレビュー
                  </label>
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
                  {preview ? (
                    <ReactMarkdown
                      remarkPlugins={[remarkGfm]}
                      rehypePlugins={[rehypeHighlight]}
                      components={markdownComponents()}
                    >
                      {preview}
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

          <button
            type='submit'
            className='bg-stone-900 dark:bg-stone-100 text-white dark:text-stone-900 px-4 py-2 rounded-lg hover:bg-stone-800 dark:hover:bg-stone-200'
          >
            作成
          </button>
        </form>
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
                remarkPlugins={[remarkGfm]}
                rehypePlugins={[rehypeHighlight]}
                components={markdownComponents()}
              >
                {preview}
              </ReactMarkdown>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
