'use client';

import ReactMarkdown from 'react-markdown';
import rehypeHighlight from 'rehype-highlight';
import rehypeSlug from 'rehype-slug';
import rehypeRaw from 'rehype-raw';
import MermaidDiagram from './MermaidDiagram';
import Callout from './Callout';
import 'highlight.js/styles/github-dark.css';
import remarkGfm from 'remark-gfm';
import { remarkCallout, preprocessCallouts } from '@/lib/remark-callout';
import type { CalloutType } from '@/lib/remark-callout';
import { useState, useRef, useCallback } from 'react';

const CALLOUT_TYPE_RE = /callout callout-(note|info|tip|warning|caution|important|warm)/;

function CodeBlock({ children, ...props }: React.ComponentProps<'pre'>) {
  const ref = useRef<HTMLPreElement>(null);
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    const text = ref.current?.textContent || '';
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, []);

  return (
    <div className='relative group'>
      <pre ref={ref} {...props}>
        {children}
      </pre>
      <button
        onClick={handleCopy}
        className='absolute top-2 right-2 p-1.5 rounded-md opacity-0 group-hover:opacity-100 transition-opacity text-xs text-gray-400 hover:text-gray-200 hover:bg-gray-700/50'
        aria-label='Copy code'
      >
        {copied ? (
          <svg className='h-4 w-4 text-green-400' viewBox='0 0 20 20' fill='currentColor'>
            <path fillRule='evenodd' d='M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z' clipRule='evenodd' />
          </svg>
        ) : (
          <svg className='h-4 w-4' viewBox='0 0 20 20' fill='currentColor'>
            <path d='M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z' />
            <path d='M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z' />
          </svg>
        )}
      </button>
    </div>
  );
}

export default function ArticleContent({ content }: { content: string }) {
  const processed = preprocessCallouts(content);

  return (
    <div className='article-content prose max-w-none dark:prose-invert'>
      <ReactMarkdown
        remarkPlugins={[remarkGfm, remarkCallout]}
        rehypePlugins={[rehypeRaw, rehypeHighlight, rehypeSlug]}
        components={{
          h2({ children, ...props }) {
            return (
              <h2 className='article-h2' {...props}>
                {children}
              </h2>
            );
          },
          a({ href, children, ...props }) {
            const isExternal = href && (href.startsWith('http://') || href.startsWith('https://'));
            return (
              <a
                href={href}
                target={isExternal ? '_blank' : undefined}
                rel={isExternal ? 'noopener noreferrer' : undefined}
                className={isExternal ? 'external-link' : ''}
                {...props}
              >
                {children}
                {isExternal && (
                  <svg
                    className='inline-block h-3.5 w-3.5 ml-0.5 -translate-y-0.5 opacity-70'
                    viewBox='0 0 20 20'
                    fill='currentColor'
                  >
                    <path d='M11 3a1 1 0 100 2h2.586l-6.293 6.293a1 1 0 101.414 1.414L15 6.414V9a1 1 0 102 0V4a1 1 0 00-1-1h-5z' />
                    <path d='M5 5a2 2 0 00-2 2v8a2 2 0 002 2h8a2 2 0 002-2v-3a1 1 0 10-2 0v3H5V7h3a1 1 0 000-2H5z' />
                  </svg>
                )}
              </a>
            );
          },
          pre({ children, ...props }) {
            return <CodeBlock {...props}>{children}</CodeBlock>;
          },
          code({ className, children, ...props }) {
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
          div({ className, children, ...props }) {
            const classStr = className || '';
            const match = classStr.match(CALLOUT_TYPE_RE);
            if (match) {
              return (
                <Callout type={match[1] as CalloutType}>{children}</Callout>
              );
            }
            return (
              <div className={className} {...props}>
                {children}
              </div>
            );
          },
          blockquote({ children, ...props }) {
            return (
              <blockquote className='article-blockquote' {...props}>
                {children}
              </blockquote>
            );
          },
          details({ children, ...props }) {
            return (
              <details
                className='my-4 rounded-lg border border-gray-200 dark:border-gray-700'
                {...props}
              >
                {children}
              </details>
            );
          },
          summary({ children, ...props }) {
            return (
              <summary
                className='cursor-pointer select-none px-4 py-2 font-semibold text-sm bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700 hover:bg-gray-200 transition-colors'
                {...props}
              >
                {children}
              </summary>
            );
          },
        }}
      >
        {processed}
      </ReactMarkdown>
    </div>
  );
}
