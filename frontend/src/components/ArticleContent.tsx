'use client';

import ReactMarkdown from 'react-markdown';
import rehypeHighlight from 'rehype-highlight';
import MermaidDiagram from './MermaidDiagram';
import 'highlight.js/styles/github.css';
import 'highlight.js/styles/github-dark.css';
import remarkGfm from 'remark-gfm';

export default function ArticleContent({ content }: { content: string }) {
  return (
    <div className='prose max-w-none dark:prose-invert'>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeHighlight]}
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
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}
