// Server Component (Node.js) は相対URLを解決できない → Gateway を直叩き
// Client Component (Browser) は Next.js rewrites で同一オリジン通信
import type { Article } from './types';

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
    body: JSON.stringify({
      question,
      model,
      api_key: apiKey,
      search_engine: searchEngine,
    }),
  });
  if (!res.ok) throw new Error('Failed to ask question');
  return res.json();
}

// agent chat
export interface AgentStep {
  thought: string;
  action: string;
  action_input: string;
  observation: string;
}

export interface AgentSource {
  article_id?: string;
  title?: string;
  url?: string;
}

export interface ChatHistoryEntry {
  role: 'user' | 'assistant';
  content: string;
}

export async function askWithAgent(
  question: string,
  model: string,
  apiKey: string,
  searchEngine: string,
  enableWebSearch: boolean,
  history: ChatHistoryEntry[],
): Promise<{ answer: string; steps: AgentStep[]; sources: AgentSource[] }> {
  const res = await fetch(`${API_BASE}/ai/agent`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      question,
      model,
      api_key: apiKey,
      search_engine: searchEngine,
      enable_web_search: enableWebSearch,
      history: JSON.stringify(history),
    }),
  });
  if (!res.ok) throw new Error('Failed to ask with agent');
  return res.json();
}

// --- Agent Streaming ---
export interface AgentStreamStepEvent {
  step_index: number;
  thought: string;
  action: string;
  action_input: string;
  observation: string;
  phase: 'llm_thinking' | 'tool_executing' | 'tool_complete';
}

export interface AgentStreamFinalEvent {
  answer: string;
  steps: AgentStep[];
  sources: AgentSource[];
}

export interface AgentStreamErrorEvent {
  message: string;
}

export interface AgentStreamCallbacks {
  onStep: (step: AgentStreamStepEvent) => void;
  onFinal: (final: AgentStreamFinalEvent) => void;
  onError: (error: string) => void;
}

/**
 * Agent ストリーミング API
 *
 * 【全体のデータの流れ】
 *
 *   fetch()           → HTTP POST リクエスト送信
 *   res.body          → ReadableStream (サーバーからのデータが少しずつ届くパイプ)
 *   getReader()       → パイプから取り出すための reader を取得
 *   read() ループ     → チャンク(Uint8Array)単位で取り出す
 *   TextDecoder       → Uint8Array → 文字列 に変換
 *   buffer            → 不完全なSSEメッセージを一時保持
 *   split("\n\n")     → SSEメッセージの境界で分割
 *   JSON.parse        → 文字列 → オブジェクト に変換
 *   callbacks.onStep  → UIにステップをリアルタイム表示
 *
 * 【なぜ buffer が必要か】
 *
 *   ネットワークはSSEメッセージの境界(\n\n)を意識してくれない。
 *   チャンクの区切りと \n\n の位置は無関係。
 *
 *   例: 2つのSSEメッセージが3チャンクに分断されて届くケース
 *
 *     チャンク1: "event: step\ndata: {\"phase\":\"llm"     ← メッセージ1の途中まで
 *     チャンク2: "_thinking\"}\n\nevent: step\ndata: {\"p" ← メッセージ1の残り + メッセージ2の途中
 *     チャンク3: "hase\":\"tool\"}\n\n"                     ← メッセージ2の残り
 *
 *   buffer に結合してから "\n\n" で分割することで、正しく2つのメッセージに分けられる。
 *
 * 【TextDecoder の stream:true がやること】
 *
 *   UTF-8 では日本語1文字が 3バイト (例: 'あ' = 0xE3 0x81 0x82)。
 *   1バイト目の先頭ビットで「この文字は何バイト必要か」がわかる:
 *     0xxxxxxx   → 1バイト (ASCII)
 *     110xxxxx   → 2バイト文字の開始 (あと1バイト必要)
 *     1110xxxx   → 3バイト文字の開始 (あと2バイト必要)
 *     11110xxx   → 4バイト文字の開始 (あと3バイト必要)
 *     10xxxxxx   → 継続バイト (文字の途中)
 *
 *   チャンク境界で分断された場合:
 *     チャンク1: [0xE3, 0x81]  ← 3バイト文字の1〜2バイト目、まだ足りない
 *     チャンク2: [0x82]         ← 3バイト目、ここで揃う → 'あ' に復号
 *
 *   stream:true なら TextDecoder が未完了バイトを内部保持し、次の decode() で続きと結合。
 *   stream:false だと未完了バイトは � になる。
 */
export async function askWithAgentStream(
  question: string,
  model: string,
  apiKey: string,
  searchEngine: string,
  enableWebSearch: boolean,
  history: ChatHistoryEntry[],
  callbacks: AgentStreamCallbacks,
): Promise<void> {
  // --- Step 1: HTTPリクエスト送信 ---
  // 通常のfetchと同じだが、サーバーは "text/event-stream" で応答するため
  // res.body が「少しずつ届くストリーム」になる
  const res = await fetch(`${API_BASE}/ai/agent/stream`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      question,
      model,
      api_key: apiKey,
      search_engine: searchEngine,
      enable_web_search: enableWebSearch,
      history: JSON.stringify(history),
    }),
  });

  if (!res.ok) {
    callbacks.onError(`HTTP ${res.status}`);
    return;
  }

  // --- Step 2: ReadableStream から reader を取得 ---
  // res.body は ReadableStream<Uint8Array>
  //   中身はまだ文字列ではない。バイト配列のストリーム。
  // getReader() すると、このストリームを独占的に読めるようになる
  //   (1つのストリームに対して reader は1つだけ)
  const reader = res.body?.getReader();
  if (!reader) {
    callbacks.onError('No response body');
    return;
  }

  // TextDecoder: Uint8Array → 文字列 の変換器
  const decoder = new TextDecoder();
  // buffer: チャンク境界で分断されたSSEメッセージを一時保持する文字列
  let buffer = '';

  try {
    // --- Step 3: チャンク受信ループ ---
    while (true) {
      // read() の戻り値:
      //   done: false, value: Uint8Array → データが届いた
      //   done: true                      → ストリーム終了(サーバーが接続を閉じた)
      // read() は次のチャンクが届くまで await でブロックする
      const { done, value } = await reader.read();
      if (done) break;

      // --- Step 4: バイト配列 → 文字列変換 ---
      // value: Uint8Array (例: [101, 118, 101, 110, 116, 58, ...])
      //   ↓ decoder.decode()
      // 文字列 (例: "event: step\ndata: {...}\n\n")
      //
      // stream:true の意味:
      //   この decode 呼び出しが「ストリームの一部」であることを伝える
      //   不完全なマルチバイト文字があれば内部バッファに保持し、
      //   次回の decode() で続きと結合して正しく復号する
      buffer += decoder.decode(value, { stream: true });

      // --- Step 5: SSEメッセージの分割 ---
      // SSEプロトコルでは "\n\n" (空行) がメッセージの区切り
      // split("\n\n") の結果:
      //   "msg1\n\nmsg2\n\nmsg3途中" → ["msg1", "msg2", "msg3途中"]
      //                                              完全↑    不完全↑
      const messages = buffer.split('\n\n');
      // pop() で最後の要素を取り出し buffer に戻す
      //   "msg3途中" は \n\n で終わっていない = 次チャンクの続きが来るはず
      //   これを buffer に保持して次のループで続きと結合
      buffer = messages.pop() || '';
      // messages に残った要素は全て \n\n で終わっていた = 完全なSSEメッセージ

      // --- Step 6: 各メッセージのパース ---
      for (const msg of messages) {
        if (!msg.trim()) continue;

        // SSEメッセージの構造:
        //   "event: step\ndata: {\"phase\":\"llm_thinking\"}"
        //    ↑ 行1          ↑ 行2
        //   行ごとに分割して event: と data: を抽出
        let eventType = '';
        let data = '';

        for (const line of msg.split('\n')) {
          // "event: step" → slice(7) → "step"
          if (line.startsWith('event: ')) {
            eventType = line.slice(7).trim();
            // "data: {...}" → slice(6) → "{...}"
          } else if (line.startsWith('data: ')) {
            data = line.slice(6);
          }
        }

        // event or data が欠けていればスキップ
        if (!eventType || !data) continue;

        try {
          // data は JSON文字列 → オブジェクトに変換
          // 例: "{\"event_type\":\"step\",\"step\":{\"phase\":\"llm_thinking\"}}"
          const parsed = JSON.parse(data);

          // イベント種別に応じて対応するコールバックを呼ぶ
          //   UI側 (ChatInterface) でリアルタイム表示に使われる
          switch (eventType) {
            case 'step':
              callbacks.onStep(parsed.step || parsed);
              break;
            case 'final_answer':
              callbacks.onFinal(parsed.final || parsed);
              break;
            case 'error':
              callbacks.onError(
                parsed.error?.message || parsed.message || 'Unknown error',
              );
              break;
          }
        } catch {
          // JSONパース失敗 = 不正なデータ。無視して次へ
        }
      }
      // ループ先頭に戻り、次のチャンクを待機
    }
  } finally {
    // ストリームのロックを解放
    // getReader() でロックされたので、使い終わったら解放が必要
    reader.releaseLock();
  }
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
      throw new Error(
        'サーバーがタイムアウトしました。AIサービスが起動しているか確認してください。',
      );
    if (res.status === 503)
      throw new Error(
        'AIサービスが利用できません。しばらくしてからお試しください。',
      );
    throw new Error(`グラフの取得に失敗しました (HTTP ${res.status})`);
  }
  const data = await res.json();
  if (!data.entities?.length)
    throw new Error('グラフデータが空です。先に記事を投稿してください。');
  return data;
}

// -- Like / Save --
export async function toggleLike(articleId: string, fingerprint: string): Promise<{ count: number; liked: boolean }> {
  const res = await fetch(`${API_BASE}/articles/${articleId}/like`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ fingerprint }),
  });
  if (!res.ok) throw new Error('Failed to toggle like');
  return res.json();
}

export async function getLikeCount(articleId: string, fingerprint: string): Promise<{ count: number; liked: boolean }> {
  const res = await fetch(`${API_BASE}/articles/${articleId}/like?fingerprint=${encodeURIComponent(fingerprint)}`);
  if (!res.ok) throw new Error('Failed to get like count');
  return res.json();
}

export async function getLikeCounts(articleIds: string[]): Promise<{ counts: { article_id: string; count: number }[] }> {
  const res = await fetch(`${API_BASE}/articles/likes`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ article_ids: articleIds }),
  });
  if (!res.ok) throw new Error('Failed to get like counts');
  return res.json();
}

export async function saveArticleBookmark(articleId: string, fingerprint: string): Promise<{ saved: boolean }> {
  const res = await fetch(`${API_BASE}/articles/${articleId}/save`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ fingerprint }),
  });
  if (!res.ok) throw new Error('Failed to save article');
  return res.json();
}

export async function unsaveArticleBookmark(articleId: string, fingerprint: string): Promise<{ unsaved: boolean }> {
  const res = await fetch(`${API_BASE}/articles/${articleId}/save`, {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ fingerprint }),
  });
  if (!res.ok) throw new Error('Failed to unsave article');
  return res.json();
}

export async function listSavedArticles(fingerprint: string): Promise<{ articles: Article[] }> {
  const res = await fetch(`${API_BASE}/articles/saved?fingerprint=${encodeURIComponent(fingerprint)}`);
  if (!res.ok) throw new Error('Failed to list saved articles');
  return res.json();
}

export async function isArticleSaved(articleId: string, fingerprint: string): Promise<{ saved: boolean }> {
  const res = await fetch(`${API_BASE}/articles/${articleId}/saved?fingerprint=${encodeURIComponent(fingerprint)}`);
  if (!res.ok) throw new Error('Failed to check saved status');
  return res.json();
}
