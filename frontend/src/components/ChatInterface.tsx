'use client';
import {
  askQuestion,
  askWithAgent,
  type AskSource,
  type AgentStep,
  type AgentSource,
  type ChatHistoryEntry,
} from '@/lib/api';
import {
  useCallback,
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import 'highlight.js/styles/github-dark.css';
import { FaRobot, FaUser } from 'react-icons/fa';
import { FiTrash2, FiHelpCircle, FiX } from 'react-icons/fi';
import { MODELS, SEARCH_ENGINES, CHAT_MODES } from '@/lib/const';
import ConfirmModal from './ConfirmModal';
import AgentSteps from './AgentSteps';
import { AiOutlineGlobal } from 'react-icons/ai';

interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
  sources?: AskSource[];
  agentSteps?: AgentStep[];
  agentSources?: AgentSource[];
}

const STORAGE_KEY_MESSAGES = 'chat_history';
const getKeyStorageKey = (modelId: string) => `ai_key_${modelId}`;

function CodeBlock({ children, ...props }: React.ComponentProps<'pre'>) {
  const ref = useRef<HTMLPreElement>(null);
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    const text = ref.current?.textContent || '';
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, []);

  return (
    <div className='relative group'>
      <pre ref={ref} {...props}>
        {children}
      </pre>
      <button
        onClick={handleCopy}
        className='absolute top-2 right-2 p-1.5 rounded-md opacity-0 group-hover:opacity-100 transition-opacity text-xs text-gray-400 hover:text-gray-200 hover:bg-gray-700/50'
        aria-label='Copy code'
      >
        {copied ? (
          <svg
            className='h-4 w-4 text-green-400'
            viewBox='0 0 20 20'
            fill='currentColor'
          >
            <path
              fillRule='evenodd'
              d='M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z'
              clipRule='evenodd'
            />
          </svg>
        ) : (
          <svg className='h-4 w-4' viewBox='0 0 20 20' fill='currentColor'>
            <path d='M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z' />
            <path d='M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z' />
          </svg>
        )}
      </button>
    </div>
  );
}

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
  const [searchEngine, setSearchEngine] = useState('bm25');
  const [chatMode, setChatMode] = useState('rag');
  const [enableWebSearch, setEnableWebSearch] = useState(false);
  const [showHelp, setShowHelp] = useState(false);

  const currentModel = MODELS.find((m) => m.id === model);
  const currentEngine = SEARCH_ENGINES.find((e) => e.id === searchEngine);

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

      if (chatMode === 'agent') {
        // 直近6件の会話履歴を構築
        const recentHistory: ChatHistoryEntry[] = messages
          .slice(-6)
          .map((m) => ({
            role: m.role,
            content: m.content,
          }));

        const { answer, steps, sources } = await askWithAgent(
          input,
          modelToSend,
          apiKey,
          searchEngine,
          enableWebSearch,
          recentHistory,
        );
        updateMessages((prev) => [
          ...prev,
          {
            role: 'assistant',
            content: answer,
            agentSteps: steps,
            agentSources: sources,
          },
        ]);
      } else {
        const { answer, sources } = await askQuestion(
          input,
          modelToSend,
          apiKey,
          searchEngine,
        );
        updateMessages((prev) => [
          ...prev,
          { role: 'assistant', content: answer, sources },
        ]);
      }
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
      {/* モデル選択 + モード切替 + 検索エンジン選択 + API Key 入力 + 履歴削除 */}
      <div className='flex items-center gap-3 p-3 border-b border-stone-300 dark:border-stone-600'>
        <select
          value={chatMode}
          onChange={(e) => {
            setChatMode(e.target.value);
            setShowHelp(false);
          }}
          className='rounded-lg border border-stone-300 dark:border-stone-500 bg-white dark:bg-stone-700 px-3 py-1.5 text-sm text-stone-900 dark:text-stone-100 focus:outline-none focus:ring-2 focus:ring-blue-500'
        >
          {CHAT_MODES.map((m) => (
            <option key={m.id} value={m.id}>
              {m.name}
            </option>
          ))}
        </select>

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

        <select
          value={searchEngine}
          onChange={(e) => setSearchEngine(e.target.value)}
          className='rounded-lg border border-stone-300 dark:border-stone-500 bg-white dark:bg-stone-700 px-3 py-1.5 text-sm text-stone-900 dark:text-stone-100 focus:outline-none focus:ring-2 focus:ring-blue-500'
        >
          {SEARCH_ENGINES.map((e) => (
            <option key={e.id} value={e.id}>
              {e.name}
            </option>
          ))}
        </select>

        {chatMode === 'agent' && (
          <button
            className={`flex items-center gap-1 text-sm text-stone-600 dark:text-stone-300 ${enableWebSearch && 'bg-blue-300 rounded-2xl dark:text-stone-800 p-1'}`}
            type='button'
            // value={enableWebSearch}
            onClick={() => setEnableWebSearch((prev) => !prev)}
          >
            <AiOutlineGlobal />
            Web検索
          </button>
        )}

        {(currentModel?.needsKey ||
          (currentEngine?.needsKey && model !== 'ollama')) && (
          <input
            type='password'
            value={apiKey}
            onChange={(e) => handleKeyChange(e.target.value)}
            placeholder='API Key（タブ閉じると消えます）'
            className='flex-1 rounded-lg border border-stone-300 dark:border-stone-500 bg-white dark:bg-stone-700 px-3 py-1.5 text-sm text-stone-900 dark:text-stone-100 placeholder-stone-400 dark:placeholder-stone-500 focus:outline-none focus:ring-2 focus:ring-blue-500'
          />
        )}

        {/* ヘルプボタン */}
        <button
          type='button'
          onClick={() => setShowHelp((prev) => !prev)}
          title='使い方を確認'
          className={`shrink-0 p-1.5 rounded-lg transition-colors ${showHelp ? 'bg-blue-100 text-blue-600 dark:bg-blue-900 dark:text-blue-300' : 'text-stone-400 hover:text-stone-600 dark:hover:text-stone-200'}`}
        >
          <FiHelpCircle size={18} />
        </button>

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

      {/* ヘルプパネル（モード別に内容を切り替え） */}
      {showHelp && (
        <div className='border-b border-stone-300 dark:border-stone-600 bg-stone-50 dark:bg-stone-800/50 p-4 text-sm text-stone-700 dark:text-stone-300'>
          <div className='flex items-center justify-between mb-2'>
            <h3 className='font-semibold text-stone-900 dark:text-stone-100'>
              {chatMode === 'agent' ? 'Agent モード' : 'RAG モード'}の使い方
            </h3>
            <button
              type='button'
              onClick={() => setShowHelp(false)}
              className='text-stone-400 hover:text-stone-600 dark:hover:text-stone-200'
            >
              <FiX size={16} />
            </button>
          </div>
          {chatMode === 'agent' ? (
            <div className='space-y-2'>
              <p>AIが自律的にツールを選択・実行して回答を導きます。</p>
              <div>
                <p className='font-medium'>
                  実行モード（モデルにより自動切替）:
                </p>
                <ul className='list-disc ml-4 mt-1 space-y-1'>
                  <li>
                    <span className='font-medium text-blue-600 dark:text-blue-400'>
                      外部モデル
                    </span>
                    （Gemini / DeepSeek / OpenAI / GLM-5）→{' '}
                    <span className='font-medium'>自律 ReAct</span>: LLMが
                    Thought → Action → Observation を自分で判断・反復
                  </li>
                  <li>
                    <span className='font-medium text-orange-600 dark:text-orange-400'>
                      Ollama
                    </span>
                    （ローカル）→{' '}
                    <span className='font-medium'>固定パイプライン</span>:
                    search → read → 回答の順序で確実に実行
                  </li>
                </ul>
              </div>
              <div>
                <p className='font-medium'>利用可能ツール:</p>
                <ul className='list-disc ml-4 mt-1 space-y-0.5'>
                  <li>
                    <code className='bg-stone-200 dark:bg-stone-700 px-1 rounded text-xs'>
                      search_wiki
                    </code>{' '}
                    Wiki内を検索
                  </li>
                  <li>
                    <code className='bg-stone-200 dark:bg-stone-700 px-1 rounded text-xs'>
                      read_article
                    </code>{' '}
                    記事全文を取得
                  </li>
                  <li>
                    <code className='bg-stone-200 dark:bg-stone-700 px-1 rounded text-xs'>
                      list_articles
                    </code>{' '}
                    記事一覧を取得
                  </li>
                  <li>
                    <code className='bg-stone-200 dark:bg-stone-700 px-1 rounded text-xs'>
                      web_search
                    </code>{' '}
                    外部サイトを検索（Web検索ON時）
                  </li>
                  <li>
                    <code className='bg-stone-200 dark:bg-stone-700 px-1 rounded text-xs'>
                      read_url
                    </code>{' '}
                    Webページ本文を取得（Web検索ON時）
                  </li>
                </ul>
              </div>
            </div>
          ) : (
            <div className='space-y-2'>
              <p>
                質問に対して Wiki 内の関連記事を検索し、その内容をもとに LLM
                が回答します。
              </p>
              <div>
                <p className='font-medium'>処理フロー:</p>
                <ol className='list-decimal ml-4 mt-1 space-y-0.5'>
                  <li>選択した検索エンジンで質問に関連する記事を検索</li>
                  <li>上位記事の全文をコンテキストとして LLM に渡す</li>
                  <li>コンテキストに基づいて回答を生成</li>
                </ol>
              </div>
              <div>
                <p className='font-medium'>検索エンジン:</p>
                <ul className='list-disc ml-4 mt-1 space-y-0.5'>
                  <li>
                    <span className='font-medium'>BM25</span>{' '}
                    キーワード一致（API Key不要）
                  </li>
                  <li>
                    <span className='font-medium'>Vector</span>{' '}
                    意味的類似度（Embedding API必要）
                  </li>
                  <li>
                    <span className='font-medium'>Hybrid</span> BM25 + Vector
                    の統合
                  </li>
                  <li>
                    <span className='font-medium'>Graph RAG</span>{' '}
                    ナレッジグラフで関連記事を横断検索
                  </li>
                </ul>
              </div>
            </div>
          )}
        </div>
      )}

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
                <div className='prose prose-sm dark:prose-invert prose-li:marker:text-stone-600 dark:prose-li:marker:text-stone-300 prose-hr:border-stone-500 dark:prose-hr:border-stone-400 prose-code:text-stone-800 dark:prose-code:text-stone-200 max-w-none'>
                  <ReactMarkdown
                    remarkPlugins={[remarkGfm]}
                    rehypePlugins={[rehypeHighlight]}
                    components={{
                      pre({ node, children, ...props }) {
                        return <CodeBlock {...props}>{children}</CodeBlock>;
                      },
                      a: ({ href, children }) => (
                        <a
                          href={href}
                          target='_blank'
                          rel='noopener noreferrer'
                        >
                          {children}
                        </a>
                      ),
                    }}
                  >
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
              {msg.agentSteps && msg.agentSteps.length > 0 && (
                <AgentSteps
                  steps={msg.agentSteps}
                  sources={msg.agentSources || []}
                />
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
