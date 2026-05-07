'use client';

import { useState } from 'react';
import ChatInterface from '@/components/ChatInterface';
import KnowledgeGraph from '@/components/KnowledgeGraph';

export default function ChatPage() {
  const [tab, setTab] = useState<'chat' | 'graph'>('chat');

  return (
    <div className='flex h-full flex-col px-3 py-3 sm:mx-auto sm:max-w-5xl sm:px-6 sm:py-4'>
      <div className='shrink-0 mb-3'>
        <div className='mb-1 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between'>
          <h1 className='text-2xl font-bold sm:text-3xl'>Chat</h1>
          <div className='flex w-full gap-1 rounded-lg bg-stone-100 p-1 dark:bg-stone-800 sm:w-auto'>
            <button
              onClick={() => setTab('chat')}
              className={`flex-1 rounded-md px-3 py-1 text-sm transition-colors sm:flex-none ${
                tab === 'chat'
                  ? 'bg-white dark:bg-stone-700 text-stone-900 dark:text-stone-100 shadow-sm'
                  : 'text-stone-500 hover:text-stone-700 dark:hover:text-stone-300'
              }`}
            >
              Chat
            </button>
            <button
              onClick={() => setTab('graph')}
              className={`flex-1 rounded-md px-3 py-1 text-sm transition-colors sm:flex-none ${
                tab === 'graph'
                  ? 'bg-white dark:bg-stone-700 text-stone-900 dark:text-stone-100 shadow-sm'
                  : 'text-stone-500 hover:text-stone-700 dark:hover:text-stone-300'
              }`}
            >
              Graph
            </button>
          </div>
        </div>
        <p className='text-md text-gray-500 dark:text-stone-400'>
          {tab === 'chat'
            ? 'Wiki の内容に基づいて AI が回答します'
            : 'ナレッジグラフのエンティティと関係を可視化します'}
        </p>
      </div>
      <div className='flex-1 min-h-0'>
        {tab === 'chat' ? <ChatInterface /> : <KnowledgeGraph />}
      </div>
    </div>
  );
}
