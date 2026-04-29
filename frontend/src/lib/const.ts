export const MODELS = [
  { id: 'ollama', name: 'Ollama（ローカル）', needsKey: false, defaultModel: 'gemma3:1b' },
  { id: 'deepseek', name: 'DeepSeek', needsKey: true, defaultModel: 'deepseek-chat' },
  { id: 'gemini', name: 'Gemini', needsKey: true, defaultModel: 'gemini-2.0-flash' },
  { id: 'glm5', name: 'GLM-5', needsKey: true, defaultModel: 'glm-5' },
  { id: 'openai', name: 'OpenAI', needsKey: true, defaultModel: 'gpt-4o-mini' },
];
