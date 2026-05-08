'use client';

import { useState } from 'react';
import ChatInterface from '@/components/ChatInterface';
import KnowledgeGraph from '@/components/KnowledgeGraph';

export default function ChatPage() {
  const [tab, setTab] = useState<'chat' | 'graph'>('chat');

  return (
    <div className='flex h-full flex-col px-3 py-3 sm:mx-auto sm:max-w-5xl sm:px-6 sm:py-4'>
      {/* タブ切り替え - 右端配置 */}
      <div className='shrink-0 mb-3 flex justify-end'>
        <div className='flex gap-1 rounded-lg bg-stone-100 p-1 dark:bg-stone-800 shadow-sm'>
          <button
            onClick={() => setTab('chat')}
            className={`rounded-md px-4 py-2 text-sm font-medium transition-colors ${
              tab === 'chat'
                ? 'bg-white dark:bg-stone-700 text-stone-900 dark:text-stone-100 shadow-sm'
                : 'text-stone-500 hover:text-stone-700 dark:hover:text-stone-300'
            }`}
          >
            Chat
          </button>
          <button
            onClick={() => setTab('graph')}
            className={`rounded-md px-4 py-2 text-sm font-medium transition-colors ${
              tab === 'graph'
                ? 'bg-white dark:bg-stone-700 text-stone-900 dark:text-stone-100 shadow-sm'
                : 'text-stone-500 hover:text-stone-700 dark:hover:text-stone-300'
            }`}
          >
            Graph
          </button>
        </div>
      </div>
      <div className='flex-1 min-h-0'>
        {tab === 'chat' ? <ChatInterface /> : <KnowledgeGraph />}
      </div>
    </div>
  );
}
