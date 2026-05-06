import MermaidDiagram from './MermaidDiagram';
import Callout from './Callout';
import type { CalloutType } from '@/lib/remark-callout';
import { useState, useRef, useCallback } from 'react';
import type { Components } from 'react-markdown';

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

// Markdownレンダリングの共通components設定
export default function markdownComponents(): Components {
  return {
    pre({ children, ...props }) {
      return <CodeBlock {...props}>{children}</CodeBlock>;
    },
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
    div({ className, children, ...props }: React.ComponentProps<'div'>) {
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
    details({ children, ...props }: React.ComponentProps<'details'>) {
      return (
        <details
          className='my-4 rounded-lg border border-gray-200 dark:border-gray-700'
          {...props}
        >
          {children}
        </details>
      );
    },
    summary({ children, ...props }: React.ComponentProps<'summary'>) {
      return (
        <summary
          className='cursor-pointer select-none px-4 py-2 font-semibold text-sm bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700 hover:bg-gray-200 transition-colors'
          {...props}
        >
          {children}
        </summary>
      );
    },
  };
}
