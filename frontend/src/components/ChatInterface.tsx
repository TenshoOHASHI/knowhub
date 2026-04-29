'use client';
import { askQuestion, type AskSource } from '@/lib/api';
import { useEffect, useRef, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { FaRobot, FaUser } from 'react-icons/fa';

interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
  sources?: AskSource[];
}

export default function ChatInterface() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);

  const messagesEndRef = useRef<HTMLDivElement>(null);

  const handleSubmit = async (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!input.trim() || loading) return;

    const useMsg: ChatMessage = { role: 'user', content: input };
    setMessages((prev) => [...prev, useMsg]);
    setInput('');
    setLoading(true);

    try {
      const { answer, sources } = await askQuestion(input);
      setMessages((prev) => [
        ...prev,
        { role: 'assistant', content: answer, sources }, // sources = [{id: "11", title:"gRPC"}]
      ]);
    } catch {
      setMessages((prev) => [
        ...prev,
        { role: 'assistant', content: 'エラーが発生しました。' },
      ]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  return (
    <div className='flex flex-col h-[calc(100vh-200px)] max-w-5xl mx-auto rounded-lg border border-stone-300 dark:border-stone-600 bg-white dark:bg-stone-800'>
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
                <div className='prose prose-sm dark:prose-invert max-w-none'>
                  <ReactMarkdown remarkPlugins={[remarkGfm]}>
                    {msg.content}
                  </ReactMarkdown>
                </div>
              ) : (
                <p>{msg.content}</p>
              )}
              {msg.sources && msg.sources.length > 0 && (
                <div className='text-xs text-stone-400 mt-2 pt-2 border-t border-stone-300 dark:border-stone-600'>
                  参照:{' '}
                  {msg.sources.map((s, idx) => (
                    <a
                      key={s.articleId + idx}
                      href={`/wiki/${s.articleId}`}
                      className='underline hover:text-stone-300 ml-1'
                    >
                      {s.title}
                      {idx < msg.sources!.length - 1 && ','}
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
    </div>
  );
}
