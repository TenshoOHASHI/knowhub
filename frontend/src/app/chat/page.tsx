import ChatInterface from '@/components/ChatInterface';

export default function ChatPage() {
  return (
    <div className='flex flex-col h-full max-w-5xl mx-auto px-6 py-4'>
      <div className='shrink-0 mb-3'>
        <h1 className='text-3xl font-bold mb-1'>Chat</h1>
        <p className='text-md text-gray-500 dark:text-stone-400'>
          Wiki の内容に基づいて AI が回答します
        </p>
      </div>
      <div className='flex-1 min-h-0'>
        <ChatInterface />
      </div>
    </div>
  );
}
