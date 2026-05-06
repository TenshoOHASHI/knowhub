import GithubSlugger from 'github-slugger';

export interface TocItem {
  id: string;
  text: string;
  level: number;
}

function toPlainHeadingText(markdownText: string): string {
  return markdownText
    .replace(/!\[([^\]]*)\]\([^)]+\)/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .replace(/`([^`]+)`/g, '$1')
    .replace(/[*_~]/g, '')
    .trim();
}

export function extractToc(content: string): TocItem[] {
  // セクションの中身を抽出：## セクション1
  const regex = /^(#{2,3})\s+(.+)$/gm;
  const items: TocItem[] = [];
  const slugger = new GithubSlugger();
  let match;
  while ((match = regex.exec(content)) !== null) {
    const level = match[1].length;
    const text = toPlainHeadingText(match[2]);
    const id = slugger.slug(text);
    items.push({ id, text, level });
  }
  return items;
}
