// Server Component (Node.js) は相対URLを解決できない → Gateway を直叩き
// Client Component (Browser) は Next.js rewrites で同一オリジン通信
const API_BASE =
  typeof window === 'undefined'
    ? process.env.SERVER_API_URL || 'http://localhost:8080/api'
    : '/api';

// -- Auth --
// ログイン状態チェック（Cookie が rewrites で Gateway に転送される）
export async function checkAuth() {
  const res = await fetch(`${API_BASE}/user/me`);
  return res.ok;
}

// -- Wiki --
export async function getArticles() {
  const res = await fetch(`${API_BASE}/articles`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch articles');
  return res.json();
}

export async function getArticle(id: string, token?: string) {
  const headers: Record<string, string> = {};
  if (token) headers['Authorization'] = `Bearer ${token}`;
  const res = await fetch(`${API_BASE}/articles/${id}`, {
    cache: 'no-store',
    headers,
  });
  if (!res.ok) throw new Error('Failed to fetch article');
  return res.json();
}

// -- Profile --
export async function getProfile() {
  const res = await fetch(`${API_BASE}/profile`);
  if (!res.ok) throw new Error('Failed to fetch profile');
  return res.json();
}

// -- Portfolio --
export async function getPortfolioItems() {
  const res = await fetch(`${API_BASE}/portfolio`);
  return res.json();
}

// -- Category --
export async function getCategories() {
  const res = await fetch(`${API_BASE}/categories`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch categories');
  return res.json();
}

export async function createCategory(data: {
  name: string;
  parent_id: string;
}) {
  const res = await fetch(`${API_BASE}/categories`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error('Failed to create category');
  return res.json();
}

export async function deleteCategory(id: string) {
  const res = await fetch(`${API_BASE}/categories/${id}`, {
    method: 'DELETE',
  });
  if (!res.ok) throw new Error('Failed to delete category');
}

// -- Article mutations --
export async function saveArticle(data: {
  id?: string;
  title: string;
  content: string;
  category_id?: string;
  visibility?: string;
}) {
  const { id, ...body } = data;
  const res = await fetch(
    id ? `${API_BASE}/articles/${id}` : `${API_BASE}/articles`,
    {
      method: id ? 'PUT' : 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    },
  );
  if (!res.ok)
    throw new Error(
      id ? 'Failed to update article' : 'Failed to create article',
    );
  return res.json();
}

export async function deleteArticle(id: string) {
  const res = await fetch(`${API_BASE}/articles/${id}`, {
    method: 'DELETE',
  });
  if (!res.ok) throw new Error('Failed to delete article');
}

// -- Profile mutation --
export async function saveProfile(data: {
  title: string;
  bio: string;
  github_url: string;
  avatar_url: string;
  twitter_url: string;
  linkedin_url: string;
  wantedly_url: string;
  skills: string;
  languages: string;
}) {
  const res = await fetch(`${API_BASE}/profile`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error('Failed to update profile');
  return res.json();
}

// -- Portfolio mutations --
export async function savePortfolioItem(data: {
  id?: string;
  title: string;
  description: string;
  url: string;
  status: string;
  category: string;
  tech_stack: string;
}) {
  const { id, ...body } = data;
  const res = await fetch(
    id ? `${API_BASE}/portfolio/${id}` : `${API_BASE}/portfolio`,
    {
      method: id ? 'PUT' : 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    },
  );
  if (!res.ok)
    throw new Error(
      id ? 'Failed to update portfolio' : 'Failed to create portfolio',
    );
  return res.json();
}

export async function deletePortfolioItem(id: string) {
  const res = await fetch(`${API_BASE}/portfolio/${id}`, {
    method: 'DELETE',
  });
  if (!res.ok) throw new Error('Failed to delete portfolio');
}

// -- Upload --
export async function uploadImage(file: File): Promise<string> {
  const formData = new FormData();
  formData.append('file', file);
  const res = await fetch(`${API_BASE}/upload`, {
    method: 'POST',
    body: formData,
  });
  if (!res.ok) throw new Error('Upload failed');
  const data = await res.json();
  return data.url;
}

// ai chat
export interface AskSource {
  article_id: string;
  title: string;
}

// model: バックエンドでプロバイダーを動的選択
// apiKey: ボディで送信（HTTPSで暗号化されるため十分安全）
export async function askQuestion(
  question: string,
  model: string,
  apiKey: string,
  searchEngine: string = '',
): Promise<{ answer: string; sources: AskSource[] }> {
  const res = await fetch(`${API_BASE}/ai/ask`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ question, model, api_key: apiKey, search_engine: searchEngine }),
  });
  if (!res.ok) throw new Error('Failed to ask question');
  return res.json();
}

// -- Knowledge Graph --
export interface EntityNode {
  id: string;
  name: string;
  type: string;
  article_ids: string[];
}

export interface RelationEdge {
  source: string;
  target: string;
  label: string;
}

export async function getKnowledgeGraph(): Promise<{
  entities: EntityNode[];
  relations: RelationEdge[];
}> {
  const res = await fetch(`${API_BASE}/ai/graph`);
  if (!res.ok) {
    if (res.status === 504 || res.status === 502)
      throw new Error('サーバーがタイムアウトしました。AIサービスが起動しているか確認してください。');
    if (res.status === 503)
      throw new Error('AIサービスが利用できません。しばらくしてからお試しください。');
    throw new Error(`グラフの取得に失敗しました (HTTP ${res.status})`);
  }
  const data = await res.json();
  if (!data.entities?.length) throw new Error('グラフデータが空です。先に記事を投稿してください。');
  return data;
}

