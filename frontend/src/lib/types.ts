// リテラルの型（読み込み専用）
export const PORTFOLIO_STATUSES = ['developing', 'completed'] as const;
// タブルから型を抽出
export type PortfolioStatus = (typeof PORTFOLIO_STATUSES)[number];

export interface PortfolioItem {
  id: string;
  title: string;
  description: string;
  url: string;
  status: PortfolioStatus;
  category: string;
  tech_stack: string; // JSON string
  created_at: { seconds: number };
  updated_at: { seconds: number };
}

export interface Article {
  id: string;
  title: string;
  content: string;
  category_id: string;
  visibility: string;
  created_at: { seconds: number };
  updated_at: { seconds: number };
}

export interface Profile {
  id: string;
  title: string;
  bio: string;
  github_url: string;
  avatar_url: string;
  twitter_url: string;
  linkedin_url: string;
  wantedly_url: string;
  skills: string; // JSON string
  languages: string; // JSON string: [{name, level}]
  created_at: { seconds: number };
  updated_at: { seconds: number };
}

export interface Category {
  id: string;
  name: string;
  parent_id: string;
}
