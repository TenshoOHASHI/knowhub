'use client';

import { useState } from 'react';
import ChatInterface from '@/components/ChatInterface';
import KnowledgeGraph from '@/components/KnowledgeGraph';

export default function ChatPage() {
  const [tab, setTab] = useState<'chat' | 'graph'>('chat');

  return (
    <div className='flex flex-col h-full max-w-5xl mx-auto px-6 py-4'>
      <div className='shrink-0 mb-3'>
        <div className='flex items-center justify-between mb-1'>
          <h1 className='text-3xl font-bold'>Chat</h1>
          <div className='flex gap-1 bg-stone-100 dark:bg-stone-800 rounded-lg p-1'>
            <button
              onClick={() => setTab('chat')}
              className={`px-3 py-1 text-sm rounded-md transition-colors ${
                tab === 'chat'
                  ? 'bg-white dark:bg-stone-700 text-stone-900 dark:text-stone-100 shadow-sm'
                  : 'text-stone-500 hover:text-stone-700 dark:hover:text-stone-300'
              }`}
            >
              Chat
            </button>
            <button
              onClick={() => setTab('graph')}
              className={`px-3 py-1 text-sm rounded-md transition-colors ${
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
