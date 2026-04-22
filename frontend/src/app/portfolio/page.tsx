import { getPortfolioItems } from '@/lib/api';
import { type PortfolioItem } from '@/lib/types';
import { FiTool, FiCheckCircle, FiExternalLink } from 'react-icons/fi';

export default async function PortfolioPage() {
  const data = await getPortfolioItems();
  const items: PortfolioItem[] = data.items;

  // statusで分ける
  const developing = items.filter((item) => item.status === 'developing');
  const completed = items.filter((item) => item.status === 'completed');

  return (
    <div className='max-w-4xl mx-auto p-6'>
      <h1 className='text-3xl font-bold mb-6'>Portfolio</h1>

      {developing.length === 0 && completed.length === 0 && (
        <p className='text-stone-400'>まだポートフォリオアイテムがありません</p>
      )}

      {/* 開発中 */}
      {developing.length > 0 && (
        <>
          <h2 className='text-xl font-semibold mb-4 flex items-center'>
            <FiTool className='mr-2' /> 開発中
          </h2>
          <div className='space-y-4 mb-8'>
            {developing.map((item) => (
              <PortfolioCard key={item.id} item={item} />
            ))}
          </div>
        </>
      )}

      {/* 完了 */}
      {completed.length > 0 && (
        <>
          <h2 className='text-xl font-semibold mb-4 flex items-center'>
            <FiCheckCircle className='mr-2' /> 完了
          </h2>
          <div className='space-y-4'>
            {completed.map((item) => (
              <PortfolioCard key={item.id} item={item} />
            ))}
          </div>
        </>
      )}
    </div>
  );
}

// 個々のポートフォリオ内容を表示
function PortfolioCard({ item }: { item: PortfolioItem }) {
  return (
    <div className='border border-black dark:border-stone-600 rounded-lg p-4 hover:shadow-md'>
      <h3 className='text-lg font-semibold'>{item.title}</h3>
      <p className='text-gray-600 dark:text-stone-400 mt-1'>
        {item.description}
      </p>
      <a
        href={item.url}
        target='_blank'
        className='inline-flex items-center text-sm mt-2 hover:underline'
      >
        <FiExternalLink className='mr-1' />
        リンク
      </a>
    </div>
  );
}
