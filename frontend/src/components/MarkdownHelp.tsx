'use client';

import { useState } from 'react';
import { FiX, FiCopy, FiCheck } from 'react-icons/fi';

const HELP_SECTIONS = [
  // ===== 基本構文 =====
  {
    title: '見出し / 太字 / 斜体',
    syntax: '# H1\n## H2\n### H3\n\n**太字** *斜体* `コード`',
  },
  {
    title: 'リスト / 引用 / 水平線',
    syntax: '- 項目1\n- 項目2\n  - ネスト\n\n> 引用文\n\n---',
  },
  {
    title: 'リンク / 画像',
    syntax: '[テキスト](URL)\n![代替テキスト](URL)',
  },
  // ===== knowhub DB設計書（Auth Service）=====
  {
    title: 'DB: users（Auth Service）',
    syntax:
      '## users table — Auth Service (:50051)\n\n| Column | Type | Nullable | Description |\n|--------|------|----------|-------------|\n| id | VARCHAR(36) | NO | PK, UUID |\n| username | VARCHAR(100) | NO | ユーザー名 |\n| email | VARCHAR(200) | NO | UNIQUE |\n| password_hash | VARCHAR(200) | NO | bcrypt済み |\n| created_at | DATETIME | NO | 作成日時 |',
  },
  // ===== knowhub DB設計書（Wiki Service）=====
  {
    title: 'DB: articles（Wiki Service）',
    syntax:
      '## articles table — Wiki Service (:50052)\n\n| Column | Type | Nullable | Description |\n|--------|------|----------|-------------|\n| id | VARCHAR(36) | NO | PK, UUID |\n| title | VARCHAR(200) | NO | 記事タイトル |\n| content | TEXT | NO | Markdown本文 |\n| created_at | DATETIME | NO | 作成日時 |\n| updated_at | DATETIME | NO | 更新日時 |',
  },
  // ===== knowhub DB設計書（Profile Service）=====
  {
    title: 'DB: profiles（Profile Service）',
    syntax:
      '## profiles table — Profile Service (:50053)\n\n| Column | Type | Nullable | Description |\n|--------|------|----------|-------------|\n| id | VARCHAR(36) | NO | PK, UUID |\n| title | VARCHAR(200) | NO | サイトタイトル |\n| bio | TEXT | NO | 自己紹介 |\n| github_url | TEXT | NO | GitHub URL |\n| created_at | DATETIME | NO | 作成日時 |\n| updated_at | DATETIME | NO | 更新日時 |',
  },
  {
    title: 'DB: portfolio_items（Profile Service）',
    syntax:
      '## portfolio_items table — Profile Service (:50053)\n\n| Column | Type | Nullable | Description |\n|--------|------|----------|-------------|\n| id | VARCHAR(36) | NO | PK, UUID |\n| title | VARCHAR(200) | NO | プロジェクト名 |\n| description | TEXT | NO | 説明 |\n| url | TEXT | NO | プロジェクトURL |\n| status | TEXT | NO | developing / completed |\n| created_at | DATETIME | NO | 作成日時 |',
  },
  // ===== API仕様書 =====
  {
    title: 'API: POST /api/articles',
    syntax:
      '## POST /api/articles\n\n| Field | Type | Required | Description |\n|-------|------|----------|-------------|\n| title | string | YES | 記事タイトル |\n| content | string | YES | Markdown本文 |\n\n**Response 201:**\n| Field | Type | Description |\n|-------|------|-------------|\n| article.id | string | UUID |\n| article.title | string | |\n| article.content | string | |\n| article.created_at | Timestamp | |\n| article.updated_at | Timestamp | |',
  },
  {
    title: 'API: GET /api/articles',
    syntax:
      '## GET /api/articles\n\n**Response 200:**\n| Field | Type | Description |\n|-------|------|-------------|\n| article[] | array | 記事一覧 |\n| article[].id | string | UUID |\n| article[].title | string | |\n| article[].content | string | |\n| article[].created_at | Timestamp | |\n| article[].updated_at | Timestamp | |',
  },
  {
    title: 'API: Auth endpoints',
    syntax:
      '## POST /api/auth/register\n| Field | Type | Required |\n|-------|------|----------|\n| username | string | YES |\n| email | string | YES |\n| password | string | YES |\n\n## POST /api/auth/login\n| Field | Type | Required |\n|-------|------|----------|\n| email | string | YES |\n| password | string | YES |',
  },
  // ===== gRPC RPC一覧 =====
  {
    title: 'gRPC: WikiServices（:50052）',
    syntax:
      '## WikiServices RPC\n\n| RPC | Request | Response | Description |\n|-----|---------|----------|-------------|\n| Create | title, content | Article | 記事作成 |\n| Get | id | Article | 記事取得 |\n| List | - | Article[] | 全記事取得 |\n| Update | id, title?, content? | Article | 部分更新 |\n| Delete | id | Empty | 記事削除 |',
  },
  {
    title: 'gRPC: AuthService（:50051）',
    syntax:
      '## AuthService RPC\n\n| RPC | Request | Response | Description |\n|-----|---------|----------|-------------|\n| Register | username, email, password | User + JWT | ユーザー登録 |\n| Login | email, password | User + JWT | ログイン |\n| VerifyToken | token | User | トークン検証 |',
  },
  {
    title: 'gRPC: ProfileService（:50053）',
    syntax:
      '## ProfileService RPC\n\n| RPC | Request | Response | Description |\n|-----|---------|----------|-------------|\n| GetProfile | - | Profile | 取得 |\n| CreateProfile | title, bio, github_url | Profile | 作成 |\n| UpdateProfile | title, bio, github_url | Profile | 更新 |\n| CreatePortfolioItem | title, desc, url, status | Item | 作成 |\n| ListPortfolioItems | - | Item[] | 一覧 |\n| UpdatePortfolioItem | id, title?, desc?, url?, status? | Item | 部分更新 |\n| DeletePortfolioItem | id | Empty | 削除 |',
  },
  // ===== コードブロック =====
  {
    title: 'Go / SQL コードブロック',
    syntax:
      '```go\nfunc main() {\n    fmt.Println("Hello")\n}\n```\n\n```sql\nSELECT * FROM articles WHERE id = ?\n```',
  },
  // ===== Mermaid図 =====
  {
    title: 'アーキテクチャ全体図',
    syntax:
      '```mermaid\ngraph TD\n  Client[Browser :3000]\n  GW[API Gateway :8080]\n  Auth[Auth Service :50051]\n  Wiki[Wiki Service :50052]\n  Prof[Profile Service :50053]\n  DB[(MySQL :3306)]\n  Cache[(Redis :6379)]\n\n  Client -->|REST| GW\n  GW -->|gRPC| Auth\n  GW -->|gRPC| Wiki\n  GW -->|gRPC| Prof\n  Auth --> DB\n  Wiki --> DB\n  Wiki --> Cache\n  Prof --> DB\n```',
  },
  {
    title: 'CQRS + キャッシュフロー',
    syntax:
      '```mermaid\nflowchart LR\n  subgraph Write[Command]\n    A[Create/Update/Delete] --> B[MySQL Write]\n    B --> C[Redis Cache Delete]\n  end\n  subgraph Read[Query]\n    D[FindAll/FindById] --> E{Redis Hit?}\n    E -->|Yes| F[Cache返却]\n    E -->|No| G[MySQL Read]\n    G --> H[Cache保存 10min]\n  end\n```',
  },
  {
    title: '記事作成シーケンス',
    syntax:
      '```mermaid\nsequenceDiagram\n  participant C as Client\n  participant GW as Gateway\n  participant W as WikiService\n  participant DB as MySQL\n  participant R as Redis\n\n  C->>GW: POST /api/articles\n  GW->>W: gRPC Create\n  W->>DB: INSERT INTO articles\n  W->>R: DEL articles:list\n  W-->>GW: Article\n  GW-->>C: 201 Created\n```',
  },
  {
    title: 'ER図（knowhub全体）',
    syntax:
      '```mermaid\nerDiagram\n  USER ||--o{ ARTICLE : writes\n  USER ||--|| PROFILE : has\n  PROFILE ||--o{ PORTFOLIO_ITEM : contains\n\n  USER {\n    string id PK\n    string username\n    string email\n    string password_hash\n    datetime created_at\n  }\n  ARTICLE {\n    string id PK\n    string title\n    text content\n    datetime created_at\n    datetime updated_at\n  }\n  PROFILE {\n    string id PK\n    string title\n    text bio\n    text github_url\n  }\n  PORTFOLIO_ITEM {\n    string id PK\n    string title\n    text description\n    text url\n    string status\n  }\n```',
  },
  {
    title: 'フローチャート',
    syntax:
      '```mermaid\nflowchart TD\n  A[リクエスト受信] --> B{認証チェック}\n  B -->|OK| C[DB書き込み]\n  B -->|NG| D[401返却]\n  C --> E[キャッシュ削除]\n  E --> F[201返却]\n```',
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
      className='absolute top-1 right-1 p-1 text-gray-400 hover:text-black dark:hover:text-stone-100 rounded'
      title='コピー'
    >
      {copied ? <FiCheck size={12} /> : <FiCopy size={12} />}
    </button>
  );
}

export default function MarkdownHelp({ isOpen, onClose }: MarkdownHelpProps) {
  if (!isOpen) return null;

  return (
    <div className='border-l border-black dark:border-stone-600 p-4 overflow-y-auto h-full thin-scrollbar'>
      <div className='flex justify-between items-center mb-4'>
        <h3 className='bg-white/80 p-2 dark:bg-stone-900/80 font-semibold text-sm'>
          Markdownリファレンス
        </h3>
        <button
          onClick={onClose}
          className='text-gray-400 hover:text-black dark:hover:text-stone-100'
        >
          <FiX size={16} />
        </button>
      </div>
      <div className='space-y-4'>
        {HELP_SECTIONS.map((section) => (
          <div key={section.title}>
            <h4 className='text-xs font-medium text-gray-500 dark:text-stone-400 mb-1'>
              {section.title}
            </h4>
            <div className='relative'>
              <pre className='text-xs bg-gray-100 dark:bg-stone-800 p-2 pr-7 rounded overflow-x-auto'>
                {section.syntax}
              </pre>
              <CopyButton text={section.syntax} />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
