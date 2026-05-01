export const MODELS = [
  {
    id: 'ollama',
    name: 'Ollama（ローカル）',
    needsKey: false,
    defaultModel: 'gemma3:1b',
  },
  {
    id: 'deepseek',
    name: 'DeepSeek',
    needsKey: true,
    defaultModel: 'deepseek-chat',
  },
  {
    id: 'gemini',
    name: 'Gemini',
    needsKey: true,
    defaultModel: 'gemini-2.0-flash',
  },
  { id: 'glm5', name: 'GLM-5', needsKey: true, defaultModel: 'glm-5' },
  { id: 'openai', name: 'OpenAI', needsKey: true, defaultModel: 'gpt-4o-mini' },
];

export const SEARCH_ENGINES = [
  // テキスト計算だけ。外部API不要。
  { id: 'bm25', name: 'BM25（キーワード検索）', needsKey: false },

  // embedding 生成に API が必要。
  // Ollama ローカルなら不要だが、4GB VPS では外部APIを使う前提。
  { id: 'vector', name: 'Vector（セマンティック検索）', needsKey: true },

  // Vector を内包するので同じく API が必要。
  { id: 'hybrid', name: 'Hybrid（BM25 + Vector）', needsKey: true },

  // エンティティ抽出に LLM API が必要。
  { id: 'graph', name: 'Graph RAG（ナレッジグラフ）', needsKey: true },
];
