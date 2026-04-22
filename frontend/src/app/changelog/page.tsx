const UPDATES = [
  {
    date: '2026-04-22',
    title: 'Markdown対応 & エディタ強化',
    changes: [
      'Markdown → HTML レンダリング（react-markdown）',
      'Mermaid図表対応（フローチャート・ER図）',
      'シンタックスハイライト（rehype-highlight）',
      'プレビュー拡大モーダル',
      'Markdownリファレンスパネル',
    ],
  },
  {
    date: '2026-04-21',
    title: 'Wikiページ改善',
    changes: [
      'サイドバー追加（カテゴリ一覧）',
      '記事検索バー',
      '戻るボタン（記事詳細ページ）',
    ],
  },
  // 過去の更新を追加...
];

export default function ChangelogPage() {
  return (
    <div className='max-w-4xl mx-auto p-6'>
      <h1 className='text-3xl font-bold mb-8'>Changelog</h1>
      <div className='space-y-8'>
        {UPDATES.map((update) => (
          <div
            key={update.date + update.title}
            className='relative pl-6 border-l-2 border-black dark:border-stone-600'
          >
            {/* 日付の丸ポチ */}
            <div className='absolute -left-2 top-1 w-4 h-4 rounded-full bg-stone-900 dark:bg-stone-100' />
            <p className='text-sm text-gray-400'>{update.date}</p>
            <h2 className='text-lg font-semibold mt-1'>{update.title}</h2>
            <ul className='mt-2 space-y-1'>
              {update.changes.map((change) => (
                <li
                  key={change}
                  className='text-sm text-gray-600 dark:text-stone-400'
                >
                  - {change}
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </div>
  );
}
