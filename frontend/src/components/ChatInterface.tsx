'use client';
import {
  askQuestion,
  askWithAgentStream,
  type AskSource,
  type AgentStep,
  type AgentSource,
  type ChatHistoryEntry,
  type AgentStreamStepEvent,
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
import { MdOutlineSupportAgent } from 'react-icons/md';

interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
  sources?: AskSource[];
  agentSteps?: AgentStep[];
  agentSources?: AgentSource[];
}

const STORAGE_KEY_MESSAGES = 'chat_history';
const getKeyStorageKey = (modelId: string) => `ai_key_${modelId}`;

/**
 * スコアをパーセンテージに正規化して表示
 * 検索エンジンごとにスコアの尺度が違うため、統一感のある表示にする
 *
 * - BM25:    0〜3  程度 → 0-100%
 * - Vector:  0〜1  （コサイン類似度） → 0-100%
 * - Hybrid:  0〜1  （正規化済み） → 0-100%
 * - Graph:   0〜N  （スコアベース） → 閾値1.0以上で参考値として表示
 * - TF-IDF: 0〜0.5程度 → 0-100%
 */
function formatSourceScore(score?: number): string | null {
  if (typeof score !== 'number' || !Number.isFinite(score)) {
    return null;
  }

  // 0-1の範囲: Vector, Hybrid（そのままパーセンテージ）
  if (score >= 0 && score <= 1) {
    return `${Math.round(score * 100)}%`;
  }

  // 1-3の範囲: BM25（最大3として正規化）
  if (score > 1 && score <= 3) {
    return `${Math.round((score / 3.0) * 100)}%`;
  }

  // 0-0.5の範囲: TF-IDF（最大0.5として正規化）
  if (score > 0 && score < 1) {
    return `${Math.round((score / 0.5) * 100)}%`;
  }

  // その他: そのまま表示（Graph RAGなど）
  return score.toFixed(2);
}

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
  const [liveSteps, setLiveSteps] = useState<AgentStreamStepEvent[]>([]);
  const [currentPhase, setCurrentPhase] = useState<string>('');
  const [currentAction, setCurrentAction] = useState<string>('');
  const [ragPhase, setRagPhase] = useState<
    'idle' | 'searching' | 'reading' | 'generating'
  >('idle');
  const abortControllerRef = useRef<AbortController | null>(null);

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

    // 前のリクエストをキャンセル
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    // 新しいAbortControllerを作成
    const abortController = new AbortController();
    abortControllerRef.current = abortController;

    updateMessages((prev) => [...prev, { role: 'user', content: input }]);
    setInput('');
    setLoading(true);
    setLiveSteps([]);
    setCurrentPhase('');
    setCurrentAction('');
    setRagPhase('searching');

    try {
      const modelToSend = currentModel?.defaultModel || model;

      if (chatMode === 'agent') {
        setRagPhase('idle');
        // 直近6件の会話履歴を構築
        const recentHistory: ChatHistoryEntry[] = messages
          .slice(-6)
          .map((m) => ({
            role: m.role,
            content: m.content,
          }));

        await askWithAgentStream(
          input,
          modelToSend,
          apiKey,
          searchEngine,
          enableWebSearch,
          recentHistory,
          {
            onStep: (step) => {
              setLiveSteps((prev) => [...prev, step]);
              setCurrentPhase(step.phase);
              setCurrentAction(step.action || '');
            },
            onFinal: (final) => {
              updateMessages((prev) => [
                ...prev,
                {
                  role: 'assistant',
                  content: final.answer,
                  agentSteps: final.steps,
                  agentSources: final.sources,
                },
              ]);
              setLiveSteps([]);
              setCurrentPhase('');
              setCurrentAction('');
            },
            onError: (error) => {
              // キャンセルによるエラーは無視
              if (error.includes('AbortError') || error.includes('キャンセル'))
                return;
              updateMessages((prev) => [
                ...prev,
                {
                  role: 'assistant',
                  content: `エラーが発生しました: ${error}`,
                },
              ]);
              setLiveSteps([]);
              setCurrentPhase('');
              setCurrentAction('');
              setRagPhase('idle');
            },
          },
          abortController.signal,
        );
      } else {
        // RAGモード: 各フェーズを表示
        setRagPhase('searching');

        // 少し遅延して「読み込み中」を表示（検索シミュレーション）
        const readingTimeout = setTimeout(() => {
          setRagPhase('reading');
        }, 800);

        // さらに遅延して「生成中」を表示（記事読み込みシミュレーション）
        const generatingTimeout = setTimeout(() => {
          setRagPhase('generating');
        }, 1600);

        try {
          const { answer, sources } = await askQuestion(
            input,
            modelToSend,
            apiKey,
            searchEngine,
          );

          // タイムアウトをクリア
          clearTimeout(readingTimeout);
          clearTimeout(generatingTimeout);

          setRagPhase('idle');
          updateMessages((prev) => [
            ...prev,
            { role: 'assistant', content: answer, sources },
          ]);
        } catch (err) {
          // エラー時もタイムアウトをクリア
          clearTimeout(readingTimeout);
          clearTimeout(generatingTimeout);
          setRagPhase('idle');
          throw err;
        }
      }
    } catch (error) {
      updateMessages((prev) => [
        ...prev,
        {
          role: 'assistant',
          content:
            error instanceof Error ? error.message : 'エラーが発生しました。',
        },
      ]);
    } finally {
      setLoading(false);
      abortControllerRef.current = null;
    }
  };

  const handleCancel = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
    setLoading(false);
    setLiveSteps([]);
    setCurrentPhase('');
    setCurrentAction('');
    setRagPhase('idle');
  };

  return (
    <div className='flex flex-col h-full rounded-xl border border-stone-200 dark:border-stone-700 bg-gradient-to-br from-white to-stone-50 dark:from-stone-900 dark:to-stone-950 shadow-lg overflow-hidden'>
      {/* タイトルヘッダー */}
      <div className='flex items-center justify-between px-5 py-4 border-b border-stone-200 dark:border-stone-700 bg-white/50 dark:bg-stone-800/50 backdrop-blur-sm'>
        <div className='flex items-center gap-3'>
          <div className='relative w-10 h-10'>
            {/* パルスアニメーション背景 */}
            <div className='absolute inset-0 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-xl animate-pulse' />
            {/* 内側のアイコン */}
            <div className='absolute inset-1 bg-white/20 rounded-lg flex items-center justify-center'>
              <MdOutlineSupportAgent
                size={18}
                className='text-white relative z-10'
              />
            </div>
            {/* 外側の波紋エフェクト */}
            <div className='absolute inset-0 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-xl animate-ping opacity-20' />
          </div>
          <div>
            <h1 className='text-lg font-bold text-stone-900 dark:text-stone-100'>
              Wiki AI
            </h1>
            <p className='text-xs text-stone-500 dark:text-stone-400'>
              Wiki の内容に基づいて AI が回答します
            </p>
          </div>
        </div>
      </div>

      {/* モデル選択 + モード切替 + 検索エンジン選択 + API Key 入力 + 履歴削除 */}
      <div className='flex flex-wrap items-center gap-2 border-b border-stone-200 p-3 dark:border-stone-700 sm:gap-3 sm:flex-nowrap sm:overflow-x-auto bg-white/30 dark:bg-stone-800/30'>
        <select
          value={chatMode}
          onChange={(e) => {
            setChatMode(e.target.value);
            setShowHelp(false);
          }}
          className='min-w-0 flex-1 basis-36 rounded-lg border border-stone-300 bg-white px-3 py-1.5 text-sm text-stone-900 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-stone-500 dark:bg-stone-700 dark:text-stone-100 sm:flex-none'
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
          className='min-w-0 flex-1 basis-36 rounded-lg border border-stone-300 bg-white px-3 py-1.5 text-sm text-stone-900 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-stone-500 dark:bg-stone-700 dark:text-stone-100 sm:flex-none'
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
          className='min-w-0 flex-1 basis-36 rounded-lg border border-stone-300 bg-white px-3 py-1.5 text-sm text-stone-900 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-stone-500 dark:bg-stone-700 dark:text-stone-100 sm:flex-none'
        >
          {SEARCH_ENGINES.map((e) => (
            <option key={e.id} value={e.id}>
              {e.name}
            </option>
          ))}
        </select>

        {chatMode === 'agent' && (
          <button
            className={`flex shrink-0 items-center justify-center gap-1 whitespace-nowrap rounded-lg px-2 py-1.5 text-sm text-stone-600 dark:text-stone-300 ${enableWebSearch && 'bg-blue-300 dark:text-stone-800'}`}
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
            className='min-w-0 flex-1 basis-full rounded-lg border border-stone-300 bg-white px-3 py-1 text-sm text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-stone-500 dark:bg-stone-700 dark:text-stone-100 dark:placeholder-stone-500 sm:basis-64'
          />
        )}

        {/* ヘルプボタン */}
        <button
          type='button'
          onClick={() => setShowHelp((prev) => !prev)}
          title='使い方を確認'
          className={`shrink-0 sm:flex sm:shrink-0 p-1.5 rounded-lg transition-colors ${showHelp ? 'bg-blue-100 text-blue-600 dark:bg-blue-900 dark:text-blue-300' : 'text-stone-400 hover:text-stone-600 dark:hover:text-stone-200'}`}
        >
          <FiHelpCircle size={18} />
        </button>

        {/* 会話履歴削除ボタン */}
        {messages.length > 0 && (
          <button
            type='button'
            onClick={() => setShowClearModal(true)}
            title='会話履歴を削除'
            className='shrink-0 sm:flex sm:shrink-0 p-1.5 text-stone-400 hover:text-red-500 transition-colors'
          >
            <FiTrash2 size={18} />
          </button>
        )}
      </div>

      {/* ヘルプパネル（モード別に内容を切り替え） */}
      {showHelp && (
        <div className='thin-scrollbar max-h-[42dvh] overflow-y-auto overscroll-contain border-b border-stone-300 bg-stone-50 p-3 text-sm text-stone-700 dark:border-stone-600 dark:bg-stone-800/50 dark:text-stone-300 sm:max-h-[50vh] sm:p-4'>
          <div className='flex items-center justify-between mb-3'>
            <h3 className='font-semibold text-base text-stone-900 dark:text-stone-100 border-l-4 border-red-400 dark:border-red-600 pl-3 bg-gradient-to-r from-red-100 via-red-50 via-red-50 to-stone-50 dark:from-red-900/30 dark:via-red-950/20 dark:via-red-950/20 dark:to-stone-800/50 py-1.5 -ml-3 pr-4 w-full'>
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
            <div className='space-y-3'>
              <p>AIが自律的にツールを選択・実行して回答を導きます。</p>

              {/* 利用制限 */}
              <div className='rounded border border-stone-300 dark:border-stone-600 bg-stone-100 dark:bg-stone-700/30 px-3 py-2'>
                <p className='font-medium text-stone-900 dark:text-stone-200 border-b-2 border-amber-400 inline-block pb-0.5'>
                  利用制限
                </p>
                <ul className='list-disc ml-4 mt-1 space-y-0.5 text-stone-700 dark:text-stone-300 text-xs'>
                  <li>
                    未ログイン利用は混雑防止のため、同時実行数と1日の利用回数に制限があります。
                  </li>
                  <li>
                    外部モデルは入力した API Key
                    を使います。利用料金・上限は各モデル提供元の設定に依存します。
                  </li>
                  <li>
                    API Key はブラウザの sessionStorage
                    に一時保存され、タブを閉じると消えます。
                  </li>
                </ul>
              </div>

              {/* 実行モード */}
              <div>
                <p className='font-medium text-stone-900 dark:text-stone-100 mb-2'>
                  実行モード（モデルにより自動切替）
                </p>

                <div className='rounded border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/20 px-3 py-2 mb-2'>
                  <div className='flex items-center gap-2 mb-1'>
                    <span className='font-semibold text-blue-900 dark:text-blue-100'>
                      外部モデル（自律 ReAct）
                    </span>
                    <span className='text-xs bg-blue-200 dark:bg-blue-800 text-blue-800 dark:text-blue-200 px-1.5 py-0.5 rounded'>
                      Gemini / DeepSeek / OpenAI / GLM-5
                    </span>
                  </div>
                  <p className='text-xs text-blue-800 dark:text-blue-200 mb-1'>
                    LLMが自分で思考（Thought）→ 行動（Action）→
                    観察（Observation）を判断・反復します。
                    複雑な質問に対して柔軟に対応できます。
                  </p>
                  <div className='text-xs text-blue-700 dark:text-blue-300 mt-1'>
                    <span className='font-medium'>おすすめ質問例:</span>
                    <ul className='list-disc ml-4 mt-0.5'>
                      <li>
                        「Goのcontextパッケージと、それに関連する記事をまとめて教えて」
                      </li>
                      <li>
                        「gRPCの実装手順をステップバイステップで説明して」
                      </li>
                      <li>「このWikiに載っている技術の全体像を整理して」</li>
                    </ul>
                  </div>
                </div>

                <div className='rounded border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/20 px-3 py-2'>
                  <div className='flex items-center gap-2 mb-1'>
                    <span className='font-semibold text-blue-900 dark:text-blue-100'>
                      Ollama（固定パイプライン）
                    </span>
                    <span className='text-xs bg-blue-200 dark:bg-blue-800 text-blue-800 dark:text-blue-200 px-1.5 py-0.5 rounded'>
                      ローカルモデル（gemma3:1b）
                    </span>
                  </div>
                  <p className='text-xs text-blue-800 dark:text-blue-200'>
                    search → read → 回答 の順序で確実に実行します。
                    シンプルで高速な回答が得られます。
                  </p>
                </div>
              </div>

              {/* 内部ツール */}
              <div>
                <p className='font-medium text-stone-900 dark:text-stone-100 mb-1'>
                  内部ツール
                </p>
                <p className='text-xs text-stone-600 dark:text-stone-400 mb-2'>
                  AI
                  エージェントが回答生成に使用する内部ツールです（一般公開はしていません）
                </p>
                <div className='grid grid-cols-1 gap-2 text-xs'>
                  <div className='flex items-center gap-3 px-3 py-2 border-l-2 border-emerald-200 dark:border-emerald-700'>
                    <code className='font-bold text-stone-900 dark:text-white text-xs'>
                      search_wiki
                    </code>
                    <span className='ml-2 text-stone-700 dark:text-stone-300'>
                      Wiki内を検索
                    </span>
                  </div>
                  <div className='flex items-center gap-3 px-3 py-2 border-l-2 border-emerald-200 dark:border-emerald-700'>
                    <code className='font-bold text-stone-900 dark:text-white text-xs'>
                      read_article
                    </code>
                    <span className='ml-2 text-stone-700 dark:text-stone-300'>
                      記事全文を取得
                    </span>
                  </div>
                  <div className='flex items-center gap-3 px-3 py-2 border-l-2 border-emerald-200 dark:border-emerald-700'>
                    <code className='font-bold text-stone-900 dark:text-white text-xs'>
                      list_articles
                    </code>
                    <span className='ml-2 text-stone-700 dark:text-stone-300'>
                      記事一覧を取得
                    </span>
                  </div>
                  <div className='flex items-center gap-3 px-3 py-2 border-l-2 border-emerald-200 dark:border-emerald-700'>
                    <code className='font-bold text-stone-900 dark:text-white text-xs'>
                      web_search
                    </code>
                    <span className='ml-2 text-stone-700 dark:text-stone-300'>
                      外部サイトを検索（Web検索ON時）
                    </span>
                  </div>
                  <div className='flex items-center gap-3 px-3 py-2 border-l-2 border-emerald-200 dark:border-emerald-700'>
                    <code className='font-bold text-stone-900 dark:text-white text-xs'>
                      read_url
                    </code>
                    <span className='ml-2 text-stone-700 dark:text-stone-300'>
                      Webページ本文を取得（Web検索ON時）
                    </span>
                  </div>
                </div>
              </div>

              {/* RAG と Agent の違い */}
              <div className='rounded border border-stone-300 dark:border-stone-600 bg-stone-100 dark:bg-stone-700/30 px-3 py-2'>
                <p className='font-medium text-stone-900 dark:text-stone-200 mb-1'>
                  RAG と Agent の違い
                </p>
                <ul className='space-y-1 text-xs text-stone-700 dark:text-stone-300'>
                  <li>
                    <span className='font-medium'>RAG:</span>{' '}
                    質問→検索→回答のワンステップ。 シンプルな質問に最適です。
                  </li>
                  <li>
                    <span className='font-medium'>Agent:</span>{' '}
                    質問を分析しながら必要な情報を収集・調査します。
                    複数の記事を比較する質問や、調査が必要な複雑な質問に最適です。
                  </li>
                </ul>
              </div>
            </div>
          ) : (
            <div className='space-y-3'>
              <p>
                質問に対して Wiki 内の関連記事を検索し、その内容をもとに LLM
                が回答します。
              </p>

              {/* 利用制限 */}
              <div className='rounded border border-stone-300 dark:border-stone-600 bg-stone-100 dark:bg-stone-700/30 px-3 py-2'>
                <p className='font-medium text-stone-900 dark:text-stone-200 border-b-2 border-amber-400 inline-block pb-0.5'>
                  利用制限
                </p>
                <ul className='list-disc ml-4 mt-1 space-y-0.5 text-stone-700 dark:text-stone-300 text-xs'>
                  <li>
                    未ログイン利用は混雑防止のため、同時実行数と1日の利用回数に制限があります。
                  </li>
                  <li>
                    Vector / Hybrid / Graph RAG は Embedding や LLM API
                    を使うため、外部モデル利用時は API Key が必要です。
                  </li>
                </ul>
              </div>

              {/* 処理フロー */}
              <div>
                <p className='font-medium text-stone-900 dark:text-stone-100 mb-2'>
                  処理フロー
                </p>
                <ol className='list-decimal ml-4 mt-1 space-y-0.5 text-xs text-stone-700 dark:text-stone-300'>
                  <li>選択した検索エンジンで質問に関連する記事を検索</li>
                  <li>上位記事の全文をコンテキストとして LLM に渡す</li>
                  <li>コンテキストに基づいて回答を生成</li>
                </ol>
              </div>

              {/* 検索エンジン選び方ガイド */}
              <div>
                <p className='font-medium text-stone-900 dark:text-stone-100 mb-2'>
                  検索エンジン選び方ガイド
                </p>

                {/* BM25 */}
                <div className='rounded border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/20 px-3 py-2 mb-2'>
                  <div className='flex items-center gap-2 mb-1'>
                    <span className='font-semibold text-blue-900 dark:text-blue-100'>
                      BM25（キーワード検索）
                    </span>
                    <span className='text-xs bg-blue-200 dark:bg-blue-800 text-blue-800 dark:text-blue-200 px-1.5 py-0.5 rounded'>
                      API Key不要
                    </span>
                  </div>
                  <p className='text-xs text-blue-800 dark:text-blue-200 mb-1'>
                    入力された単語が記事に含まれているかを検索します。
                    専門用語や固有名詞の検索に適しています。
                  </p>
                  <div className='text-xs text-blue-700 dark:text-blue-300'>
                    <span className='font-medium'>おすすめ質問例:</span>
                    <ul className='list-disc ml-4 mt-0.5'>
                      <li>「Go言語のcontextパッケージについて」</li>
                      <li>「gRPCの使い方を教えて」</li>
                      <li>「Docker Composeの書き方は？」</li>
                    </ul>
                  </div>
                </div>

                {/* Vector */}
                <div className='rounded border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/20 px-3 py-2 mb-2'>
                  <div className='flex items-center gap-2 mb-1'>
                    <span className='font-semibold text-blue-900 dark:text-blue-100'>
                      Vector（意味検索）
                    </span>
                    <span className='text-xs bg-blue-200 dark:bg-blue-800 text-blue-800 dark:text-blue-200 px-1.5 py-0.5 rounded'>
                      API Key必要
                    </span>
                  </div>
                  <p className='text-xs text-blue-800 dark:text-blue-200 mb-1'>
                    記事の「意味」を理解して検索します。
                    同じ単語を使わなくても似た概念の記事を見つけられます。
                  </p>
                  <div className='text-xs text-blue-700 dark:text-blue-300'>
                    <span className='font-medium'>おすすめ質問例:</span>
                    <ul className='list-disc ml-4 mt-0.5'>
                      <li>「非同期処理の実装方法」</li>
                      <li>「マイクロサービス間の通信手段」</li>
                      <li>「データベースのパフォーマンス改善」</li>
                    </ul>
                  </div>
                </div>

                {/* Hybrid */}
                <div className='rounded border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/20 px-3 py-2 mb-2'>
                  <div className='flex items-center gap-2 mb-1'>
                    <span className='font-semibold text-blue-900 dark:text-blue-100'>
                      Hybrid（複合検索）
                    </span>
                    <span className='text-xs bg-blue-200 dark:bg-blue-800 text-blue-800 dark:text-blue-200 px-1.5 py-0.5 rounded'>
                      API Key必要
                    </span>
                  </div>
                  <p className='text-xs text-blue-800 dark:text-blue-200 mb-1'>
                    BM25とVectorを組み合わせた最もバランスの良い検索です。
                    どの検索エンジンにするか迷ったらこれを選んでください。
                  </p>
                  <div className='text-xs text-blue-700 dark:text-blue-300'>
                    <span className='font-medium'>おすすめ質問例:</span>
                    <ul className='list-disc ml-4 mt-0.5'>
                      <li>「GoでAPIサーバーを立てたい」</li>
                      <li>「Reactの状態管理ベストプラクティス」</li>
                      <li>「JWT認証の実装手順」</li>
                    </ul>
                  </div>
                </div>

                {/* Graph RAG */}
                <div className='rounded border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/20 px-3 py-2'>
                  <div className='flex items-center gap-2 mb-1'>
                    <span className='font-semibold text-blue-900 dark:text-blue-100'>
                      Graph RAG（ナレッジグラフ）
                    </span>
                    <span className='text-xs bg-blue-200 dark:bg-blue-800 text-blue-800 dark:text-blue-200 px-1.5 py-0.5 rounded'>
                      API Key必要
                    </span>
                  </div>
                  <p className='text-xs text-blue-800 dark:text-blue-200 mb-1'>
                    記事同士の関連性を分析したナレッジグラフを使って検索します。
                    複数のトピックにまたがる質問に強いです。
                  </p>
                  <div className='text-xs text-blue-700 dark:text-blue-300'>
                    <span className='font-medium'>おすすめ質問例:</span>
                    <ul className='list-disc ml-4 mt-0.5'>
                      <li>「Goの学習ロードマップを組みたい」</li>
                      <li>「Web開発全体の流れを知りたい」</li>
                      <li>「バックエンドエンジニアになるには？」</li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* メッセージ一覧 */}
      <div className='flex-1 overflow-y-auto space-y-4 p-4 bg-gradient-to-b from-transparent to-stone-50/50 dark:to-stone-900/30'>
        {messages.length === 0 && (
          <div className='flex flex-col items-center justify-center h-full text-stone-400 gap-3'>
            <div className='w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500/20 to-indigo-600/20 dark:from-blue-500/10 dark:to-indigo-600/10 flex items-center justify-center'>
              <FaRobot size={32} className='text-blue-500 dark:text-blue-400' />
            </div>
            <p className='text-sm font-medium'>Wikiについて質問してください</p>
          </div>
        )}
        {messages.map((msg, i) => (
          <div
            key={i}
            className={`flex min-w-0 gap-2 ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
          >
            {msg.role === 'assistant' && (
              <div className='shrink-0 w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center shadow-md'>
                <FaRobot size={14} className='text-white' />
              </div>
            )}
            <div
              className={`min-w-0 max-w-[85%] rounded-2xl px-4 py-2.5 text-sm sm:max-w-[80%] ${
                msg.role === 'user'
                  ? 'bg-gradient-to-br from-blue-600 to-blue-700 text-white shadow-md'
                  : 'bg-white dark:bg-stone-800 text-stone-900 dark:text-stone-100 border border-stone-200 dark:border-stone-700 shadow-sm'
              }`}
            >
              {msg.role === 'assistant' ? (
                <div className='prose prose-sm dark:prose-invert prose-li:marker:text-stone-600 dark:prose-li:marker:text-stone-300 prose-hr:border-stone-500 dark:prose-hr:border-stone-400 prose-code:text-stone-800 dark:prose-code:text-stone-200 max-w-none'>
                  <ReactMarkdown
                    remarkPlugins={[remarkGfm]}
                    rehypePlugins={[rehypeHighlight]}
                    components={{
                      pre({ children, ...props }) {
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
                  <div className='mb-1 font-medium'>参照記事</div>
                  <div className='flex flex-wrap gap-1.5'>
                    {msg.sources.map((s) => {
                      const score = formatSourceScore(s.relevance_score);
                      return (
                        <a
                          key={s.article_id}
                          href={`/wiki/${s.article_id}`}
                          className='inline-flex max-w-full items-center gap-1 rounded border border-stone-300 px-2 py-1 hover:border-blue-400 hover:text-blue-500 dark:border-stone-500 dark:hover:border-blue-400 dark:hover:text-blue-300'
                        >
                          <span className='truncate'>{s.title}</span>
                          {score && (
                            <span className='shrink-0 rounded bg-stone-200 px-1.5 py-0.5 text-[10px] text-stone-700 dark:bg-stone-700 dark:text-stone-200'>
                              score {score}
                            </span>
                          )}
                        </a>
                      );
                    })}
                  </div>
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
            <div className='shrink-0 w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center shadow-md'>
              <FaRobot size={14} className='text-white' />
            </div>
            <div className='bg-white dark:bg-stone-800 rounded-2xl px-4 py-2.5 text-sm text-stone-400 relative pr-10 min-w-[200px] border border-stone-200 dark:border-stone-700 shadow-sm'>
              {chatMode === 'agent' ? (
                <div className='flex flex-col gap-1'>
                  {/* Wiki内を検索中 */}
                  <div className='flex items-center gap-2'>
                    <div
                      className={`w-2 h-2 rounded-full ${currentAction === 'search_wiki' && currentPhase === 'tool_executing' ? 'bg-blue-500 animate-pulse' : currentPhase === 'tool_complete' && liveSteps.some((s) => s.action === 'search_wiki') ? 'bg-emerald-500' : 'bg-stone-400'}`}
                    />
                    <span
                      className={
                        (currentAction === 'search_wiki' &&
                          currentPhase === 'tool_executing') ||
                        (currentPhase === 'tool_complete' &&
                          liveSteps.some((s) => s.action === 'search_wiki'))
                          ? 'text-stone-600 dark:text-stone-300'
                          : ''
                      }
                    >
                      Wiki内を検索中...
                    </span>
                  </div>
                  {/* 記事を読み込んでいます */}
                  <div className='flex items-center gap-2'>
                    <div
                      className={`w-2 h-2 rounded-full ${currentAction === 'read_article' && currentPhase === 'tool_executing' ? 'bg-blue-500 animate-pulse' : currentPhase === 'tool_complete' && liveSteps.some((s) => s.action === 'read_article') ? 'bg-emerald-500' : 'bg-stone-400'}`}
                    />
                    <span
                      className={
                        (currentAction === 'read_article' &&
                          currentPhase === 'tool_executing') ||
                        (currentPhase === 'tool_complete' &&
                          liveSteps.some((s) => s.action === 'read_article'))
                          ? 'text-stone-600 dark:text-stone-300'
                          : ''
                      }
                    >
                      関連記事を読み込んでいます...
                    </span>
                  </div>
                  {/* Web検索中（有効時のみ表示） */}
                  {enableWebSearch && (
                    <div className='flex items-center gap-2'>
                      <div
                        className={`w-2 h-2 rounded-full ${currentAction === 'web_search' && currentPhase === 'tool_executing' ? 'bg-amber-500 animate-pulse' : currentPhase === 'tool_complete' && liveSteps.some((s) => s.action === 'web_search') ? 'bg-emerald-500' : 'bg-stone-400'}`}
                      />
                      <span
                        className={
                          (currentAction === 'web_search' &&
                            currentPhase === 'tool_executing') ||
                          (currentPhase === 'tool_complete' &&
                            liveSteps.some((s) => s.action === 'web_search'))
                            ? 'text-stone-600 dark:text-stone-300'
                            : ''
                        }
                      >
                        Webを検索中...
                      </span>
                    </div>
                  )}
                  {/* URLを読み込んでいます（有効時のみ表示） */}
                  {enableWebSearch && (
                    <div className='flex items-center gap-2'>
                      <div
                        className={`w-2 h-2 rounded-full ${currentAction === 'read_url' && currentPhase === 'tool_executing' ? 'bg-amber-500 animate-pulse' : currentPhase === 'tool_complete' && liveSteps.some((s) => s.action === 'read_url') ? 'bg-emerald-500' : 'bg-stone-400'}`}
                      />
                      <span
                        className={
                          (currentAction === 'read_url' &&
                            currentPhase === 'tool_executing') ||
                          (currentPhase === 'tool_complete' &&
                            liveSteps.some((s) => s.action === 'read_url'))
                            ? 'text-stone-600 dark:text-stone-300'
                            : ''
                        }
                      >
                        ページを読み込んでいます...
                      </span>
                    </div>
                  )}
                  {/* 回答を生成しています */}
                  <div className='flex items-center gap-2'>
                    <div
                      className={`w-2 h-2 rounded-full ${currentPhase === 'llm_thinking' && liveSteps.length > 0 ? 'bg-blue-500 animate-pulse' : 'bg-stone-400'}`}
                    />
                    <span
                      className={
                        currentPhase === 'llm_thinking' && liveSteps.length > 0
                          ? 'text-stone-600 dark:text-stone-300'
                          : ''
                      }
                    >
                      回答を生成しています...
                    </span>
                  </div>
                </div>
              ) : chatMode === 'rag' && ragPhase !== 'idle' ? (
                <div className='flex flex-col gap-1'>
                  <div className='flex items-center gap-2'>
                    <div
                      className={`w-2 h-2 rounded-full ${ragPhase === 'searching' ? 'bg-blue-500 animate-pulse' : 'bg-stone-400'}`}
                    />
                    <span
                      className={
                        ragPhase === 'searching'
                          ? 'text-stone-600 dark:text-stone-300'
                          : ''
                      }
                    >
                      Wiki内を検索中...
                    </span>
                  </div>
                  <div className='flex items-center gap-2'>
                    <div
                      className={`w-2 h-2 rounded-full ${ragPhase === 'reading' ? 'bg-blue-500 animate-pulse' : 'bg-stone-400'}`}
                    />
                    <span
                      className={
                        ragPhase === 'reading'
                          ? 'text-stone-600 dark:text-stone-300'
                          : ''
                      }
                    >
                      関連記事を読み込んでいます...
                    </span>
                  </div>
                  <div className='flex items-center gap-2'>
                    <div
                      className={`w-2 h-2 rounded-full ${ragPhase === 'generating' ? 'bg-blue-500 animate-pulse' : 'bg-stone-400'}`}
                    />
                    <span
                      className={
                        ragPhase === 'generating'
                          ? 'text-stone-600 dark:text-stone-300'
                          : ''
                      }
                    >
                      回答を生成しています...
                    </span>
                  </div>
                </div>
              ) : (
                '回答中...'
              )}
              {liveSteps.length > 0 && (
                <div className='mt-2'>
                  <AgentSteps
                    steps={liveSteps.map((s) => ({
                      thought: s.thought || '',
                      action: s.action || '',
                      action_input: s.action_input || '',
                      observation: s.observation || '',
                    }))}
                    sources={[]}
                  />
                </div>
              )}
              <button
                type='button'
                onClick={handleCancel}
                className='absolute top-2 right-2 shrink-0 rounded p-1 text-stone-500 hover:text-stone-700 hover:bg-stone-300 dark:text-stone-400 dark:hover:text-stone-200 dark:hover:bg-stone-600 transition-colors'
                title='キャンセル'
              >
                <FiX size={14} />
              </button>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* 入力フォーム */}
      <form
        onSubmit={handleSubmit}
        className='flex gap-2 border-t border-stone-200 dark:border-stone-700 p-3 bg-white/50 dark:bg-stone-800/50 backdrop-blur-sm sm:p-4'
      >
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder='Wikiについて質問してください...'
          disabled={loading}
          className='min-w-0 flex-1 rounded-xl border border-stone-300 dark:border-stone-600 bg-white dark:bg-stone-900 px-4 py-2.5 text-sm text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50 dark:text-stone-100 dark:placeholder-stone-500 shadow-sm'
        />
        <button
          type='submit'
          disabled={loading}
          className='shrink-0 rounded-xl bg-gradient-to-r from-blue-600 to-blue-700 px-5 py-2.5 text-sm font-medium text-white hover:from-blue-700 hover:to-blue-800 disabled:cursor-not-allowed disabled:opacity-50 shadow-md transition-all'
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
