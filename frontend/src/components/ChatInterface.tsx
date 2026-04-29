'use client';
import { askQuestion, type AskSource } from '@/lib/api';
import {
  useCallback,
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { FaRobot, FaUser } from 'react-icons/fa';
import { FiTrash2 } from 'react-icons/fi';
import { MODELS } from '@/lib/const';
import ConfirmModal from './ConfirmModal';

interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
  sources?: AskSource[];
}

const STORAGE_KEY_MESSAGES = 'chat_history';
const getKeyStorageKey = (modelId: string) => `ai_key_${modelId}`;

// --- useSyncExternalStore: localStorage を React 19 対応で安全に読み書きする ---
// localStorage 変更を通知するための購読機構
const chatListeners = new Set<() => void>();
function subscribeChat(listener: () => void) {
  chatListeners.add(listener);
  return () => {
    chatListeners.delete(listener);
  };
}
// キャッシュ: 同じデータなら同じ配列参照を返す（新配列を返すと無限ループ）
const emptyMessages: ChatMessage[] = [];
let chatCache: ChatMessage[] = emptyMessages;
let chatCacheRaw: string | null = null;

function getChatSnapshot(): ChatMessage[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY_MESSAGES);
    if (raw === chatCacheRaw) return chatCache;
    chatCacheRaw = raw;
    chatCache = raw ? JSON.parse(raw) : emptyMessages;
    return chatCache;
  } catch {
    return emptyMessages;
  }
}

function getChatServerSnapshot(): ChatMessage[] {
  return emptyMessages;
}

function emitChatChange() {
  localStorage.getItem(STORAGE_KEY_MESSAGES); // trigger re-read
  chatListeners.forEach((l) => l());
}

export default function ChatInterface() {
  // useSyncExternalStore: SSR では空配列、CSR では localStorage を読む
  // setState in useEffect 不要 → React 19 警告なし + Hydrationエラーなし
  const messages = useSyncExternalStore(
    subscribeChat,
    getChatSnapshot,
    getChatServerSnapshot,
  );
  const updateMessages = useCallback(
    (updater: ChatMessage[] | ((prev: ChatMessage[]) => ChatMessage[])) => {
      const current = getChatSnapshot();
      const next = typeof updater === 'function' ? updater(current) : updater;
      localStorage.setItem(STORAGE_KEY_MESSAGES, JSON.stringify(next));
      emitChatChange(); // React に再読み込みを通知
    },
    [],
  );

  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);

  const [model, setModel] = useState('ollama');
  const [apiKey, setApiKey] = useState('');

  const currentModel = MODELS.find((m) => m.id === model);

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const [showClearModal, setShowClearModal] = useState(false);

  // API Key の初期化は不要:
  // 初期モデルは ollama（needsKey: false）なので入力欄が非表示
  // モデル切替時に handleModelChange で sessionStorage から読み込む

  const handleModelChange = (newModel: string) => {
    setModel(newModel);
    const savedKey = sessionStorage.getItem(getKeyStorageKey(newModel)) || '';
    setApiKey(savedKey);
  };

  const handleKeyChange = (key: string) => {
    setApiKey(key);
    sessionStorage.setItem(getKeyStorageKey(model), key);
  };

  // 新しいメッセージが追加されたらスクロール
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleClearHistory = () => {
    localStorage.removeItem(STORAGE_KEY_MESSAGES);
    setShowClearModal(false);
    emitChatChange();
  };

  const handleSubmit = async (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!input.trim() || loading) return;

    if (currentModel?.needsKey && !apiKey.trim()) {
      updateMessages((prev) => [
        ...prev,
        { role: 'user', content: input },
        { role: 'assistant', content: 'API Key を入力してください。' },
      ]);
      setInput('');
      return;
    }

    updateMessages((prev) => [...prev, { role: 'user', content: input }]);
    setInput('');
    setLoading(true);

    try {
      const modelToSend = currentModel?.defaultModel || model;
      const { answer, sources } = await askQuestion(input, modelToSend, apiKey);
      updateMessages((prev) => [
        ...prev,
        { role: 'assistant', content: answer, sources },
      ]);
    } catch {
      updateMessages((prev) => [
        ...prev,
        { role: 'assistant', content: 'エラーが発生しました。' },
      ]);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className='flex flex-col h-full rounded-lg border border-stone-300 dark:border-stone-600 bg-white dark:bg-stone-800'>
      {/* モデル選択 + API Key 入力 + 履歴削除 */}
      <div className='flex items-center gap-3 p-3 border-b border-stone-300 dark:border-stone-600'>
        <select
          value={model}
          onChange={(e) => handleModelChange(e.target.value)}
          className='rounded-lg border border-stone-300 dark:border-stone-500 bg-white dark:bg-stone-700 px-3 py-1.5 text-sm text-stone-900 dark:text-stone-100 focus:outline-none focus:ring-2 focus:ring-blue-500'
        >
          {MODELS.map((m) => (
            <option key={m.id} value={m.id}>
              {m.name}
            </option>
          ))}
        </select>

        {currentModel?.needsKey && (
          <input
            type='password'
            value={apiKey}
            onChange={(e) => handleKeyChange(e.target.value)}
            placeholder='API Key（タブ閉じると消えます）'
            className='flex-1 rounded-lg border border-stone-300 dark:border-stone-500 bg-white dark:bg-stone-700 px-3 py-1.5 text-sm text-stone-900 dark:text-stone-100 placeholder-stone-400 dark:placeholder-stone-500 focus:outline-none focus:ring-2 focus:ring-blue-500'
          />
        )}

        {/* 会話履歴削除ボタン */}
        {messages.length > 0 && (
          <button
            type='button'
            onClick={() => setShowClearModal(true)}
            title='会話履歴を削除'
            className='shrink-0 p-1.5 text-stone-400 hover:text-red-500 transition-colors'
          >
            <FiTrash2 size={18} />
          </button>
        )}
      </div>

      {/* メッセージ一覧 */}
      <div className='flex-1 overflow-y-auto space-y-4 p-4'>
        {messages.length === 0 && (
          <div className='flex flex-col items-center justify-center h-full text-stone-400 gap-2'>
            <FaRobot size={40} />
            <p className='text-sm'>Wikiについて質問してください</p>
          </div>
        )}
        {messages.map((msg, i) => (
          <div
            key={i}
            className={`flex gap-2 ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
          >
            {msg.role === 'assistant' && (
              <div className='shrink-0 w-8 h-8 rounded-full bg-stone-300 dark:bg-stone-600 flex items-center justify-center'>
                <FaRobot
                  size={16}
                  className='text-stone-600 dark:text-stone-300'
                />
              </div>
            )}
            <div
              className={`rounded-lg px-4 py-2 max-w-[80%] text-sm ${
                msg.role === 'user'
                  ? 'bg-blue-600 text-white'
                  : 'bg-stone-200 dark:bg-stone-700 text-stone-900 dark:text-stone-100'
              }`}
            >
              {msg.role === 'assistant' ? (
                <div className='prose prose-sm dark:prose-invert prose-li:marker:text-stone-600 dark:prose-li:marker:text-stone-300 prose-hr:border-stone-500 dark:prose-hr:border-stone-400 prose-pre:bg-stone-600 dark:prose-pre:bg-stone-900 prose-code:text-stone-800 dark:prose-code:text-stone-200 max-w-none'>
                  <ReactMarkdown remarkPlugins={[remarkGfm]}>
                    {msg.content}
                  </ReactMarkdown>
                </div>
              ) : (
                <p>{msg.content}</p>
              )}
              {msg.sources && msg.sources.length > 0 && (
                <div className='text-xs text-stone-500 dark:text-stone-400 mt-2 pt-2 border-t border-stone-300 dark:border-stone-500'>
                  参照:{' '}
                  {msg.sources.map((s) => (
                    <a
                      key={s.article_id}
                      href={`/wiki/${s.article_id}`}
                      className='underline hover:text-blue-400 ml-1'
                    >
                      {s.title}
                    </a>
                  ))}
                </div>
              )}
            </div>
            {msg.role === 'user' && (
              <div className='shrink-0 w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center'>
                <FaUser size={14} className='text-white' />
              </div>
            )}
          </div>
        ))}
        {loading && (
          <div className='flex gap-2 justify-start'>
            <div className='shrink-0 w-8 h-8 rounded-full bg-stone-300 dark:bg-stone-600 flex items-center justify-center'>
              <FaRobot
                size={16}
                className='text-stone-600 dark:text-stone-300'
              />
            </div>
            <div className='bg-stone-200 dark:bg-stone-700 rounded-lg px-4 py-2 text-sm text-stone-400'>
              回答中...
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* 入力フォーム */}
      <form
        onSubmit={handleSubmit}
        className='flex gap-2 p-4 border-t border-stone-300 dark:border-stone-600'
      >
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder='Wikiについて質問してください...'
          disabled={loading}
          className='flex-1 rounded-lg border border-stone-300 dark:border-stone-500 bg-white dark:bg-stone-700 px-3 py-2 text-sm text-stone-900 dark:text-stone-100 placeholder-stone-400 dark:placeholder-stone-500 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50'
        />
        <button
          type='submit'
          disabled={loading}
          className='rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed'
        >
          送信
        </button>
      </form>

      {/* 履歴削除確認モーダル・他のボタンでモーダル表示を発火させる */}

      <ConfirmModal
        isOpen={showClearModal}
        title='会話履歴の削除'
        message='すべての会話履歴を削除しますか？この操作は元に戻せません。'
        onConfirm={handleClearHistory}
        onCancel={() => setShowClearModal(false)}
      />
    </div>
  );
}
