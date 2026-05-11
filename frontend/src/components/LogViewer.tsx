'use client';

import { useState, useEffect, useRef, useCallback } from 'react';
import {
  FiPlay,
  FiPause,
  FiArrowDown,
  FiArrowUp,
  FiTrash2,
  FiRefreshCw,
} from 'react-icons/fi';

// サービス定義（色分け用）
// 本番: ai, gateway 等のアプリサービスが Docker で動作
// 開発: db, cache, searxng のみ Docker で動作（アプリは go run で直接起動）
const PROD_SERVICES = [
  { id: 'ai', label: 'AI', dot: 'bg-blue-400', tag: 'bg-blue-400/20 text-blue-300' },
  { id: 'gateway', label: 'Gateway', dot: 'bg-emerald-400', tag: 'bg-emerald-400/20 text-emerald-300' },
  { id: 'auth', label: 'Auth', dot: 'bg-purple-400', tag: 'bg-purple-400/20 text-purple-300' },
  { id: 'wiki', label: 'Wiki', dot: 'bg-amber-400', tag: 'bg-amber-400/20 text-amber-300' },
  { id: 'nginx', label: 'Nginx', dot: 'bg-rose-400', tag: 'bg-rose-400/20 text-rose-300' },
  { id: 'frontend', label: 'Frontend', dot: 'bg-cyan-400', tag: 'bg-cyan-400/20 text-cyan-300' },
] as const;

const DEV_SERVICES = [
  { id: 'db', label: 'MySQL', dot: 'bg-blue-400', tag: 'bg-blue-400/20 text-blue-300' },
  { id: 'cache', label: 'Redis', dot: 'bg-rose-400', tag: 'bg-rose-400/20 text-rose-300' },
  { id: 'searxng', label: 'SearXNG', dot: 'bg-amber-400', tag: 'bg-amber-400/20 text-amber-300' },
] as const;

const SERVICES =
  process.env.NODE_ENV === 'production' ? PROD_SERVICES : DEV_SERVICES;

const LOG_LEVELS = ['all', 'info', 'warn', 'error'] as const;
type LogLevel = (typeof LOG_LEVELS)[number];

interface LogEntry {
  id: number;
  raw: string;
  level: string;
  service: string;
  timestamp: string;
  message: string;
}

// ログ行をパースする
function parseLogLine(raw: string): Omit<LogEntry, 'id'> {
  // docker compose logs のプレフィックス: "service-name  | ..."
  let service = '';
  let logContent = raw;

  const pipeIndex = raw.indexOf(' | ');
  if (pipeIndex !== -1) {
    // "ai-1  | {json...}" のようなフォーマット
    const prefix = raw.substring(0, pipeIndex).trim();
    // サービス名を抽出（"ai-1" → "ai"）
    service = prefix.replace(/-\d+$/, '');
    logContent = raw.substring(pipeIndex + 3);
  }

  // slog JSON フォーマットの解析を試みる
  let level = '';
  let timestamp = '';
  let message = logContent;

  try {
    const parsed = JSON.parse(logContent);
    level = (parsed.level || '').toUpperCase();
    timestamp = parsed.time || '';
    message = parsed.msg || logContent;
    if (parsed.service) {
      service = parsed.service.toLowerCase();
    }
  } catch {
    // JSON でない場合はそのまま
    // レベルを文字列マッチで推定
    // slog: "level":"ERROR"
    // MySQL: [Warning], [ERROR], [Note], [System]
    // SearXNG: ERROR:, WARNING:
    // Redis: # (warning/error), * (notice/info)
    const lc = logContent;
    if (
      lc.includes('"level":"ERROR"') ||
      lc.includes('[ERROR]') ||
      /\bERROR:/.test(lc)
    ) {
      level = 'ERROR';
    } else if (
      lc.includes('"level":"WARN"') ||
      lc.includes('[WARN]') ||
      lc.includes('[Warning]') ||
      /\bWARNING:/.test(lc)
    ) {
      level = 'WARN';
    } else if (
      lc.includes('"level":"INFO"') ||
      lc.includes('[INFO]') ||
      lc.includes('[Note]') ||
      lc.includes('[System]') ||
      /^\d+:[A-Z] .+ \* /.test(lc) // Redis: "1:M ... * message"
    ) {
      level = 'INFO';
    } else if (
      lc.includes('"level":"DEBUG"') ||
      lc.includes('[DEBUG]')
    ) {
      level = 'DEBUG';
    } else if (/^\d+:[A-Z] .+ # /.test(lc)) {
      // Redis: "1:M ... # warning message"
      level = 'WARN';
    }
  }

  return { raw, level, service, timestamp, message };
}

// ログレベルの色
function levelColor(level: string): string {
  switch (level) {
    case 'ERROR':
      return 'text-red-400';
    case 'WARN':
      return 'text-amber-400';
    case 'INFO':
      return 'text-stone-200';
    case 'DEBUG':
      return 'text-stone-400';
    default:
      return 'text-stone-200';
  }
}

// レベルバッジの色
function levelBadgeColor(level: string): string {
  switch (level) {
    case 'ERROR':
      return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400';
    case 'WARN':
      return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400';
    case 'INFO':
      return 'bg-stone-100 text-stone-600 dark:bg-stone-700 dark:text-stone-400';
    case 'DEBUG':
      return 'bg-stone-50 text-stone-500 dark:bg-stone-800 dark:text-stone-500';
    default:
      return 'bg-stone-100 text-stone-600 dark:bg-stone-700 dark:text-stone-400';
  }
}

// サービスの色を取得
function serviceColor(serviceName: string): string {
  const svc = SERVICES.find((s) => serviceName.includes(s.id));
  return svc?.tag || 'bg-stone-400/20 text-stone-300';
}

// 通信量のフォーマット
function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

// 月次通信量のストレージキー
function monthlyKey(): string {
  const now = new Date();
  return `logs_bytes_${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
}

const MAX_LOGS = 500;
const API_BASE = '/api';

export default function LogViewer() {
  const [selectedServices, setSelectedServices] = useState<string[]>([]);
  const [logLevel, setLogLevel] = useState<LogLevel>('all');
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isStreaming, setIsStreaming] = useState(false);
  const [autoScroll, setAutoScroll] = useState(true);
  const [bytesReceived, setBytesReceived] = useState(0);
  const [monthlyBytes, setMonthlyBytes] = useState(() => {
    if (typeof window === 'undefined') return 0;
    const stored = localStorage.getItem(monthlyKey());
    return stored ? parseInt(stored, 10) : 0;
  });
  const [connectionStatus, setConnectionStatus] = useState<
    'disconnected' | 'connecting' | 'connected'
  >('disconnected');

  const logEndRef = useRef<HTMLDivElement>(null);
  const eventSourceRef = useRef<EventSource | null>(null);
  const logIdRef = useRef(0);

  // オートスクロール
  useEffect(() => {
    if (autoScroll && logEndRef.current) {
      logEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [logs, autoScroll]);

  // SSE 接続の切断
  const disconnect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    setIsStreaming(false);
    setConnectionStatus('disconnected');
  }, []);

  // SSE 接続
  const connect = useCallback(() => {
    disconnect();

    const params = new URLSearchParams();
    if (selectedServices.length > 0) {
      params.set('services', selectedServices.join(','));
    }
    if (logLevel !== 'all') {
      params.set('level', logLevel);
    }

    const url = `${API_BASE}/logs/stream?${params.toString()}`;
    setConnectionStatus('connecting');

    const es = new EventSource(url);
    eventSourceRef.current = es;

    es.onopen = () => {
      setConnectionStatus('connected');
      setIsStreaming(true);
    };

    es.addEventListener('log', (event: MessageEvent) => {
      const size = new Blob([event.data]).size;
      setBytesReceived((prev) => prev + size);
      setMonthlyBytes((prev) => {
        const newTotal = prev + size;
        localStorage.setItem(monthlyKey(), String(newTotal));
        return newTotal;
      });

      const parsed = parseLogLine(event.data);
      const entry: LogEntry = {
        id: logIdRef.current++,
        ...parsed,
      };

      setLogs((prev) => {
        const next = [...prev, entry];
        // リングバッファ: 最大500件
        if (next.length > MAX_LOGS) {
          return next.slice(next.length - MAX_LOGS);
        }
        return next;
      });
    });

    es.addEventListener('error', (event: MessageEvent) => {
      console.error('SSE error event:', event);
    });

    es.onerror = () => {
      setConnectionStatus('disconnected');
      setIsStreaming(false);
    };
  }, [selectedServices, logLevel, disconnect]);

  // タブ非アクティブ時の自動切断
  useEffect(() => {
    const handleVisibility = () => {
      if (document.hidden && isStreaming) {
        disconnect();
      }
    };
    document.addEventListener('visibilitychange', handleVisibility);
    return () => {
      document.removeEventListener('visibilitychange', handleVisibility);
    };
  }, [isStreaming, disconnect]);

  // コンポーネントのアンマウント時にクリーンアップ
  useEffect(() => {
    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
    };
  }, []);

  // サービス選択切り替え
  const toggleService = (serviceId: string) => {
    setSelectedServices((prev) =>
      prev.includes(serviceId)
        ? prev.filter((s) => s !== serviceId)
        : [...prev, serviceId],
    );
  };

  // ログクリア
  const clearLogs = () => {
    setLogs([]);
    logIdRef.current = 0;
  };

  // Docker アクション実行
  const execAction = async (service: string, action: string) => {
    try {
      const res = await fetch(`${API_BASE}/logs/action`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ service, action }),
      });
      const data = await res.json();
      if (!res.ok) {
        alert(`Action failed: ${data.error || 'unknown error'}`);
      }
    } catch (err) {
      alert(`Action failed: ${err}`);
    }
  };

  // タイムスタンプのフォーマット
  const formatTimestamp = (ts: string): string => {
    if (!ts) return '';
    try {
      const d = new Date(ts);
      return d.toLocaleTimeString('ja-JP', {
        hour12: false,
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });
    } catch {
      return ts;
    }
  };

  // 接続ステータスの色
  const statusColor =
    connectionStatus === 'connected'
      ? 'bg-emerald-500'
      : connectionStatus === 'connecting'
        ? 'bg-amber-500 animate-pulse'
        : 'bg-stone-400';

  return (
    <div className='space-y-4'>
      {/* コントロールパネル */}
      <div className='rounded-xl border border-stone-200 dark:border-stone-700 bg-white dark:bg-stone-800 shadow-sm p-4 space-y-3'>
        {/* サービス選択 */}
        <div className='flex flex-wrap items-center gap-2'>
          <span className='text-xs font-medium text-stone-500 dark:text-stone-400 w-16'>
            サービス
          </span>
          {SERVICES.map((svc) => (
            <button
              key={svc.id}
              onClick={() => toggleService(svc.id)}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                selectedServices.includes(svc.id)
                  ? 'bg-stone-900 text-white dark:bg-stone-100 dark:text-stone-900'
                  : 'bg-stone-100 text-stone-600 hover:bg-stone-200 dark:bg-stone-700 dark:text-stone-400 dark:hover:bg-stone-600'
              }`}
            >
              <span className={`w-2 h-2 rounded-full ${svc.dot}`} />
              {svc.label}
            </button>
          ))}
        </div>

        {/* レベル + コントロール */}
        <div className='flex flex-wrap items-center gap-3'>
          <div className='flex items-center gap-2'>
            <span className='text-xs font-medium text-stone-500 dark:text-stone-400 w-16'>
              レベル
            </span>
            <select
              value={logLevel}
              onChange={(e) => setLogLevel(e.target.value as LogLevel)}
              className='px-3 py-1.5 rounded-lg text-xs border border-stone-200 dark:border-stone-600 bg-white dark:bg-stone-700 text-stone-900 dark:text-stone-100'
            >
              {LOG_LEVELS.map((l) => (
                <option key={l} value={l}>
                  {l.toUpperCase()}
                </option>
              ))}
            </select>
          </div>

          <div className='flex items-center gap-2 ml-auto'>
            {/* 接続ステータス */}
            <div className='flex items-center gap-1.5'>
              <span className={`w-2 h-2 rounded-full ${statusColor}`} />
              <span className='text-xs text-stone-500 dark:text-stone-400'>
                {connectionStatus === 'connected'
                  ? '接続中'
                  : connectionStatus === 'connecting'
                    ? '接続中...'
                    : '未接続'}
              </span>
            </div>

            {/* 通信量 */}
            <span className='text-xs text-stone-400 dark:text-stone-500 px-2'>
              今月 {formatBytes(monthlyBytes)}
              {bytesReceived > 0 && ` (今回 ${formatBytes(bytesReceived)})`}
            </span>

            {/* ストリーミング ON/OFF */}
            {isStreaming ? (
              <button
                onClick={disconnect}
                className='flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-red-100 text-red-700 hover:bg-red-200 dark:bg-red-900/30 dark:text-red-400 dark:hover:bg-red-900/50 transition-all'
              >
                <FiPause size={12} />
                停止
              </button>
            ) : (
              <button
                onClick={connect}
                className='flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-emerald-100 text-emerald-700 hover:bg-emerald-200 dark:bg-emerald-900/30 dark:text-emerald-400 dark:hover:bg-emerald-900/50 transition-all'
              >
                <FiPlay size={12} />
                開始
              </button>
            )}

            {/* オートスクロール */}
            <button
              onClick={() => setAutoScroll(!autoScroll)}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                autoScroll
                  ? 'bg-stone-900 text-white dark:bg-stone-100 dark:text-stone-900'
                  : 'bg-stone-100 text-stone-600 dark:bg-stone-700 dark:text-stone-400'
              }`}
              title={
                autoScroll ? 'オートスクロール ON' : 'オートスクロール OFF'
              }
            >
              {autoScroll ? <FiArrowDown size={12} /> : <FiArrowUp size={12} />}
            </button>

            {/* ログクリア */}
            <button
              onClick={clearLogs}
              className='flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-stone-100 text-stone-600 hover:bg-stone-200 dark:bg-stone-700 dark:text-stone-400 dark:hover:bg-stone-600 transition-all'
              title='ログクリア'
            >
              <FiTrash2 size={12} />
            </button>
          </div>
        </div>

        {/* Docker アクション（本番のみ） */}
        {process.env.NODE_ENV === 'production' && (
          <div className='flex flex-wrap items-center gap-2 pt-2 border-t border-stone-100 dark:border-stone-700'>
            <span className='text-xs font-medium text-stone-500 dark:text-stone-400 w-16'>
              操作
            </span>
            <button
              onClick={() => execAction('nginx', 'reload')}
              className='flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-stone-100 text-stone-600 hover:bg-stone-200 dark:bg-stone-700 dark:text-stone-400 dark:hover:bg-stone-600 transition-all'
            >
              <FiRefreshCw size={12} />
              Nginx Reload
            </button>
          </div>
        )}
      </div>

      {/* ログ表示エリア */}
      <div className='rounded-xl border border-stone-200 dark:border-stone-700 bg-stone-950 shadow-sm overflow-hidden'>
        <div className='h-[500px] overflow-y-auto p-4 font-mono text-sm leading-relaxed'>
          {logs.length === 0 ? (
            <div className='flex items-center justify-center h-full text-stone-500'>
              {isStreaming
                ? 'ログを待機中...'
                : 'ストリーミングを開始してください'}
            </div>
          ) : (
            logs.map((entry) => (
              <div
                key={entry.id}
                className='flex gap-2 py-0.5 hover:bg-stone-900/50'
              >
                {/* タイムスタンプ */}
                <span className='text-stone-400 shrink-0 w-20'>
                  {formatTimestamp(entry.timestamp)}
                </span>
                {/* サービスタグ */}
                {entry.service && (
                  <span
                    className={`shrink-0 px-2 py-0.5 rounded text-xs font-medium ${serviceColor(entry.service)}`}
                  >
                    {entry.service.toUpperCase()}
                  </span>
                )}
                {/* レベルバッジ */}
                {entry.level && (
                  <span
                    className={`shrink-0 px-2 py-0.5 rounded text-xs font-medium ${levelBadgeColor(entry.level)}`}
                  >
                    {entry.level}
                  </span>
                )}
                {/* メッセージ */}
                <span className={`break-all ${levelColor(entry.level)}`}>
                  {entry.message}
                </span>
              </div>
            ))
          )}
          <div ref={logEndRef} />
        </div>

        {/* フッター */}
        <div className='flex items-center justify-between px-3 py-1.5 border-t border-stone-800 bg-stone-900'>
          <span className='text-xs text-stone-400'>
            {logs.length} 行 / {MAX_LOGS} max
          </span>
          <span className='text-xs text-stone-400'>
            {selectedServices.length === 0
              ? '全サービス'
              : selectedServices.join(', ')}{' '}
            | {logLevel.toUpperCase()}
          </span>
        </div>
      </div>
    </div>
  );
}
