import { getArticle } from '@/lib/api';
import Link from 'next/link';
import { FiArrowLeft } from 'react-icons/fi';
import ArticleContent from '@/components/ArticleContent';

interface Props {
  params: Promise<{ id: string }>;
}

export default async function ArticleDetailPage({ params }: Props) {
  const { id } = await params;
  const data = await getArticle(id);

  const article = data.Article;

  return (
    <div className='max-w-4xl mx-auto p-6'>
      <Link
        href='/wiki'
        className='inline-flex items-center text-md text-gray-600 dark:text-stone-400 hover:text-black dark:hover:text-stone-100 mb-6'
      >
        <FiArrowLeft className='mr-1' />
        Wikiに戻る
      </Link>
      <h1 className='text-3xl font-bold mb-4'>{article.title}</h1>
      <p className='text-stone-400 mb-6'>
        {new Date(article.created_at.seconds * 1000).toLocaleDateString(
          'ja-JP',
        )}
      </p>
      <ArticleContent content={article.content} />
    </div>
  );
}
