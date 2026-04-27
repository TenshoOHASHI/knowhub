export interface TocItem {
  id: string;
  text: string;
  level: number;
}

// Match rehype-slug / github-slugger behavior
function slug(text: string): string {
  // 英数字・漢字・ハイフン以外を全部消す
  return (
    text
      // 小文字に変換
      .toLowerCase()
      // 前後の空白除去
      .trim()
      // [\s] と \s はまったく同じ動作
      .replace(/[\s]+/g, '-') // \s = スペース・タブ・改行・全角スペース など, + 1回以上がないと、１とつずつ繋がってしまう
      // CJK統合漢字（漢字・一部の日本語）の範囲、\w,英数字とアンダースコア [a-z0-9_]
      .replace(/[^\w\u4e00-\u9faf-]/g, '') // カッコ内以外の文字を削除
  );
}

export function extractToc(content: string): TocItem[] {
  // セクションの中身を抽出：## セッション１
  const regex = /^(#{2,3})\s+(.+)$/gm;
  const items: TocItem[] = [];
  // keyに使用
  const countMap: Record<string, number> = {};
  let match;
  while ((match = regex.exec(content)) !== null) {
    // # | ## => 1 or 2
    const level = match[1].length;
    // \sセッション => セッション
    const text = match[2].trim();
    const baseId = slug(text);
    const count = countMap[baseId] || 0;
    countMap[baseId] = count + 1;
    const id = count > 0 ? `${baseId}-${count}` : baseId;
    items.push({ id, text, level });
  }
  return items;
}
