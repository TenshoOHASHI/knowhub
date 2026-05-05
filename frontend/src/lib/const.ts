export const MODELS = [
  {
    id: 'ollama',
    name: 'Ollama（gemma3:1b）',
    needsKey: false,
    defaultModel: 'gemma3:1b',
  },
  {
    id: 'deepseek',
    name: 'DeepSeek(v4-pro)',
    needsKey: true,
    defaultModel: 'deepseek-v4-pro',
  },
  {
    id: 'gemini',
    name: 'Gemini(2.0-flash)',
    needsKey: true,
    defaultModel: 'gemini-2.0-flash',
  },
  { id: 'glm5', name: 'GLM-5.1', needsKey: true, defaultModel: 'glm-5.1' },
  {
    id: 'openai',
    name: 'OpenAI(4o-mini)',
    needsKey: true,
    defaultModel: 'gpt-4o-mini',
  },
];

export const SEARCH_ENGINES = [
  // テキスト計算だけ。外部API不要。
  { id: 'bm25', name: 'BM25（キーワード検索）', needsKey: false },

  // embedding 生成に API が必要。
  // Ollama ローカルなら不要だが、4GB VPS では外部APIを使う前提。
  { id: 'vector', name: 'Vector（意味検索）', needsKey: true },

  // Vector を内包するので同じく API が必要。
  { id: 'hybrid', name: 'Hybrid（複合検索）', needsKey: true },

  // エンティティ抽出に LLM API が必要。
  { id: 'graph', name: 'Graph RAG（ナレッジグラフ）', needsKey: true },
];

export const CHAT_MODES = [
  { id: 'rag', name: 'RAG' },
  { id: 'agent', name: 'Agent' },
];
