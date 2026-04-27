import MermaidDiagram from './MermaidDiagram';

// Markdownレンダリングの共通components設定
export default function markdownComponents() {
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
