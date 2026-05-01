'use client';

import ReactMarkdown from 'react-markdown';
import rehypeHighlight from 'rehype-highlight';
import rehypeSlug from 'rehype-slug';
import rehypeRaw from 'rehype-raw';
import MermaidDiagram from './MermaidDiagram';
import Callout from './Callout';
import 'highlight.js/styles/github.css';
import 'highlight.js/styles/github-dark.css';
import remarkGfm from 'remark-gfm';
import { remarkCallout, preprocessCallouts } from '@/lib/remark-callout';
import type { CalloutType } from '@/lib/remark-callout';

const CALLOUT_TYPE_RE = /callout callout-(note|info|tip|warning|caution|important|warm)/;

export default function ArticleContent({ content }: { content: string }) {
  const processed = preprocessCallouts(content);

  return (
    <div className='prose max-w-none dark:prose-invert'>
      <ReactMarkdown
        remarkPlugins={[remarkGfm, remarkCallout]}
        rehypePlugins={[rehypeRaw, rehypeHighlight, rehypeSlug]}
        components={{
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
          div({ node, className, children, ...props }) {
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
