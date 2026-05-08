'use client';

import { useState, useEffect } from 'react';
import { FiX, FiCopy, FiCheck, FiChevronDown, FiChevronRight } from 'react-icons/fi';

const HELP_SECTIONS = [
  // ===== 基本構文 =====
  {
    category: '基本構文',
    items: [
      {
        title: '見出し / 太字 / 斜体',
        syntax: '# H1\n## H2\n### H3\n\n**太字** *斜体* `コード`',
        description: '見出しは#の数でレベルを表現',
      },
      {
        title: 'リスト / 引用 / 水平線',
        syntax: '- 項目1\n- 項目2\n  - ネスト\n\n> 引用文\n\n---',
        description: '箇条書き、引用ブロック、区切り線',
      },
      {
        title: 'リンク / 画像',
        syntax: '[テキスト](URL)\n![代替テキスト](URL)',
        description: 'リンクと埋め込み画像',
      },
    ],
  },
  // ===== コードブロック =====
  {
    category: 'コードブロック',
    items: [
      {
        title: 'シンタックスハイライト',
        syntax: '```go\nfunc main() {\n    fmt.Println("Hello")\n}\n```\n\n```sql\nSELECT * FROM articles WHERE id = ?\n```',
        description: '言語指定でハイライト（go, sql, typescript, javascript, bash など）',
      },
    ],
  },
  // ===== コールアウト & 折りたたみ =====
  {
    category: 'コールアウト & 折りたたみ',
    items: [
      {
        title: 'コールアウト（Zenn風）',
        syntax: ':::message\nメモ内容\n:::\n\n:::message info\n情報\n:::\n\n:::message tip\nヒント\n:::\n\n:::message warning\n警告\n:::\n\n:::message alert\n注意\n:::\n\n:::message warm\n温かみのある色\n:::',
        description: 'Zenn記法のコールアウト',
      },
      {
        title: 'コールアウト（GitHub風）',
        syntax: '> [!NOTE]\n> メモ\n\n> [!TIP]\n> ヒント\n\n> [!WARNING]\n> 警告\n\n> [!CAUTION]\n> 注意\n\n> [!IMPORTANT]\n> 重要',
        description: 'GitHub記法のコールアウト',
      },
      {
        title: '折りたたみブロック',
        syntax: '<details>\n<summary>クリックして展開</summary>\n\n```go\nfmt.Println("hello")\n```\n\n</details>',
        description: 'クリックで展開/折りたたみ',
      },
    ],
  },
  // ===== Mermaid図 =====
  {
    category: 'Mermaid図',
    items: [
      {
        title: 'アーキテクチャ全体図',
        syntax: '```mermaid\ngraph TD\n  Client[Browser :3000]\n  GW[API Gateway :8080]\n  Auth[Auth Service :50051]\n  Wiki[Wiki Service :50052]\n  Prof[Profile Service :50053]\n  DB[(MySQL :3306)]\n  Cache[(Redis :6379)]\n\n  Client -->|REST| GW\n  GW -->|gRPC| Auth\n  GW -->|gRPC| Wiki\n  GW -->|gRPC| Prof\n  Auth --> DB\n  Wiki --> DB\n  Wiki --> Cache\n  Prof --> DB\n```',
        description: 'サービス間の接続関係',
      },
      {
        title: 'CQRS + キャッシュフロー',
        syntax: '```mermaid\nflowchart LR\n  subgraph Write[Command]\n    A[Create/Delivery/Delete] --> B[MySQL Write]\n    B --> C[Redis Cache Delete]\n  end\n  subgraph Read[Query]\n    D[FindAll/FindById] --> E{Redis Hit?}\n    E -->|Yes| F[Cache返却]\n    E -->|No| G[MySQL Read]\n    G --> H[Cache保存 10min]\n  end\n```',
        description: 'CQRSパターンでの読み書き分離',
      },
      {
        title: 'シーケンス図',
        syntax: '```mermaid\nsequenceDiagram\n  participant C as Client\n  participant GW as Gateway\n  participant W as WikiService\n  participant DB as MySQL\n\n  C->>GW: POST /api/articles\n  GW->>W: gRPC Create\n  W->>DB: INSERT INTO articles\n  GW-->>C: 201 Created\n```',
        description: '記事作成のシーケンス',
      },
    ],
  },
  // ===== DB設計 =====
  {
    category: 'DB設計',
    items: [
      {
        title: 'users（Auth Service）',
        syntax: '| Column | Type | Nullable | Description |\n|--------|------|----------|-------------|\n| id | VARCHAR(36) | NO | PK, UUID |\n| username | VARCHAR(100) | NO | ユーザー名 |\n| email | VARCHAR(200) | NO | UNIQUE |\n| password_hash | VARCHAR(200) | NO | bcrypt済み |',
        description: 'ユーザーテーブル（:50051）',
      },
      {
        title: 'articles（Wiki Service）',
        syntax: '| Column | Type | Nullable | Description |\n|--------|------|----------|-------------|\n| id | VARCHAR(36) | NO | PK, UUID |\n| title | VARCHAR(200) | NO | 記事タイトル |\n| content | TEXT | NO | Markdown本文 |\n| created_at | DATETIME | NO | 作成日時 |\n| updated_at | DATETIME | NO | 更新日時 |',
        description: '記事テーブル（:50052）',
      },
    ],
  },
  // ===== API =====
  {
    category: 'APIエンドポイント',
    items: [
      {
        title: 'POST /api/articles',
        syntax: '| Field | Type | Required | Description |\n|-------|------|----------|-------------|\n| title | string | YES | 記事タイトル |\n| content | string | YES | Markdown本文 |',
        description: '記事作成',
      },
      {
        title: 'Auth: POST /api/auth/register',
        syntax: '| Field | Type | Required |\n|-------|------|----------|\n| username | string | YES |\n| email | string | YES |\n| password | string | YES |',
        description: 'ユーザー登録',
      },
    ],
  },
  // ===== gRPC =====
  {
    category: 'gRPCサービス',
    items: [
      {
        title: 'WikiService（:50052）',
        syntax: '| RPC | Request | Response | Description |\n|-----|---------|----------|-------------|\n| Create | title, content | Article | 記事作成 |\n| Get | id | Article | 記事取得 |\n| List | - | Article[] | 全記事取得 |',
        description: 'Wiki gRPCインターフェース',
      },
      {
        title: 'AuthService（:50051）',
        syntax: '| RPC | Request | Response | Description |\n|-----|---------|----------|-------------|\n| Register | username, email, password | User + JWT | ユーザー登録 |\n| Login | email, password | User + JWT | ログイン |',
        description: '認証 gRPCインターフェース',
      },
    ],
  },
  // ===== フロントエンドパターン =====
  {
    category: 'フロントエンドパターン',
    items: [
      {
        title: 'AbortController によるキャンセル処理',
        syntax: `// キャンセル機能実装パターン
const abortControllerRef = useRef<AbortController | null>(null);

// リクエスト開始時に AbortController を作成
const abortController = new AbortController();
abortControllerRef.current = abortController;

// fetch に signal を渡す
await fetch(url, { signal: abortController.signal });

// キャンセル処理
const handleCancel = () => {
  abortControllerRef.current?.abort();
  abortControllerRef.current = null;
};`,
        description: '非同期リクエストのキャンセル処理',
      },
      {
        title: 'コールバックパターン',
        syntax: `// コールバック関数を含むオブジェクトを渡すパターン
askWithAgentStream(
  question,
  model,
  apiKey,
  searchEngine,
  enableWebSearch,
  history,
  {
    onStep: (step) => { /* ステップ実行時 */ },
    onFinal: (final) => { /* 完了時 */ },
    onError: (error) => { /* エラー時 */ },
  },
  abortController.signal,
);`,
        description: 'ストリーミング処理でのイベント通知パターン',
      },
      {
        title: 'useState + useRef の使い分け',
        syntax: `// useState: UI状態管理（再レンダーが必要）
const [loading, setLoading] = useState(false);

// useRef: 再レンダー不要な値の保持
const abortControllerRef = useRef<AbortController | null>(null);
const messagesEndRef = useRef<HTMLDivElement>(null);`,
        description: 'フックの使い分け',
      },
    ],
  },
  // ===== コンポーネント設計 =====
  {
    category: 'コンポーネント設計',
    items: [
      {
        title: 'カードコンポーネント',
        syntax: `// 基本構造
<div className='rounded-lg border border-stone-200 dark:border-stone-600 p-4 bg-white dark:bg-stone-800'>
  <h3 className='text-lg font-semibold mb-2'>カードタイトル</h3>
  <p className='text-sm text-stone-600 dark:text-stone-400'>コンテンツ</p>
</div>

// ホバー効果
className='hover:bg-stone-50 dark:hover:bg-stone-700/50 transition-colors'`,
        description: 'カード型コンポーネント',
      },
      {
        title: 'ボタンスタイル',
        syntax: `// プライマリボタン
className='px-4 py-2 rounded-lg bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50'

// セカンダリボタン
className='px-4 py-2 rounded-lg border border-stone-300 dark:border-stone-600 hover:bg-stone-100 dark:hover:bg-stone-800'

// 危険ボタン
className='px-4 py-2 rounded-lg bg-red-500 text-white hover:bg-red-600'`,
        description: 'ボタンのバリエーション',
      },
      {
        title: 'フォーム入力欄',
        syntax: `// 標準入力欄
<input
  className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-4 py-3 bg-white dark:bg-stone-800 text-base focus:outline-none focus:ring-2 focus:ring-blue-500'
  placeholder='入力してください...'
/>

// セレクトボックス
<select
  className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-4 py-3 bg-white dark:bg-stone-800'
>
  <option value=''>選択してください</option>
</select>`,
        description: 'フォーム要素のスタイリング',
      },
    ],
  },
];

interface MarkdownHelpProps {
  isOpen: boolean;
  onClose: () => void;
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <button
      type='button'
      onClick={handleCopy}
      className='absolute top-2 right-2 p-1.5 text-stone-400 hover:text-stone-600 dark:hover:text-stone-200 bg-stone-100 dark:bg-stone-800 rounded transition-colors'
      title='コピー'
    >
      {copied ? <FiCheck size={14} /> : <FiCopy size={14} />}
    </button>
  );
}

export default function MarkdownHelp({ isOpen, onClose }: MarkdownHelpProps) {
  const [openCategories, setOpenCategories] = useState<Record<string, boolean>>({});

  // ESCキーで閉じる
  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handleEsc);
    return () => window.removeEventListener('keydown', handleEsc);
  }, [onClose]);

  if (!isOpen) return null;

  const toggleCategory = (category: string) => {
    setOpenCategories((prev) => ({ ...prev, [category]: !prev[category] }));
  };

  return (
    <div
      className='fixed inset-0 z-50 bg-black/40 backdrop-blur-sm flex items-center justify-center p-4'
      onClick={onClose}
    >
      <div
        className='bg-white dark:bg-stone-900 rounded-xl shadow-2xl w-full max-w-3xl max-h-[85vh] overflow-hidden flex flex-col'
        onClick={(e) => e.stopPropagation()}
      >
        {/* ヘッダー */}
        <div className='flex items-center justify-between px-6 py-4 border-b border-stone-200 dark:border-stone-700 bg-stone-50 dark:bg-stone-800/50'>
          <h2 className='text-lg font-bold text-stone-900 dark:text-stone-100'>
            Markdownリファレンス & 設計パターン
          </h2>
          <button
            onClick={onClose}
            className='text-stone-400 hover:text-stone-600 dark:hover:text-stone-200 transition-colors p-1 rounded hover:bg-stone-200 dark:hover:bg-stone-700'
            title='閉じる (ESC)'
          >
            <FiX size={20} />
          </button>
        </div>

        {/* コンテンツ */}
        <div className='flex-1 overflow-y-auto p-6 thin-scrollbar'>
          <div className='space-y-4'>
            {HELP_SECTIONS.map((section) => (
              <div key={section.category}>
                <button
                  type='button'
                  onClick={() => toggleCategory(section.category)}
                  className='flex items-center gap-2 w-full text-left font-semibold text-stone-800 dark:text-stone-200 hover:bg-stone-200/50 dark:hover:bg-stone-800/30 rounded px-3 py-2 transition-colors'
                >
                  {openCategories[section.category] ? <FiChevronDown size={16} /> : <FiChevronRight size={16} />}
                  {section.category}
                </button>

                {openCategories[section.category] && (
                  <div className='mt-3 ml-4 space-y-4'>
                    {section.items.map((item) => (
                      <div key={item.title} className='group'>
                        <div className='flex items-center justify-between mb-1'>
                          <h4 className='text-sm font-medium text-stone-700 dark:text-stone-300'>
                            {item.title}
                          </h4>
                          {item.description && (
                            <span className='text-xs text-stone-500 dark:text-stone-400'>{item.description}</span>
                          )}
                        </div>
                        <div className='relative'>
                          <pre className='text-xs bg-stone-100 dark:bg-stone-900 border border-stone-300 dark:border-stone-700 p-3 pr-10 rounded overflow-x-auto'>
                            {item.syntax}
                          </pre>
                          <CopyButton text={item.syntax} />
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* フッター */}
        <div className='px-6 py-3 border-t border-stone-200 dark:border-stone-700 bg-stone-50 dark:bg-stone-800/50 text-xs text-stone-500 dark:text-stone-400 text-center'>
          ESC キーまたは背景クリックで閉じる
        </div>
      </div>
    </div>
  );
}
