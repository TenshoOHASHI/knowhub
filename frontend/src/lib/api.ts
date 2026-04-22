const API_BASE = 'http://localhost:8080/api';

// -- wiki --
export async function getArticles() {
  const res = await fetch(`${API_BASE}/articles`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch articles');
  return res.json();
}

export async function getArticle(id: string) {
  const res = await fetch(
    `${API_BASE}/articles/${id}`,
    // next: { revalidate: 60 }, // 60秒間キャッシュ、期限切れたら再取得
    { cache: 'no-store' },
  );

  if (!res.ok) throw new Error('Failed to fetch article');

  return res.json();
}

export async function createArticle(data: { title: string; content: string }) {
  const res = await fetch(`${API_BASE}/articles`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error('Failed to create article');

  return res.json();
}

// -- Profile --
export async function getProfile() {
  const res = await fetch(`${API_BASE}/profile`);

  if (!res.ok) throw new Error('Failed to fetch article');
  return res.json();
}

// -- Portfolio
export async function getPortfolioItems() {
  const res = await fetch(`${API_BASE}/portfolio`);
  return res.json();
}
