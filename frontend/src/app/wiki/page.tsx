import WikiClient from '@/components/WikiClient';
import { getArticles } from '@/lib/api';
import type { Article } from '@/lib/types';

export default async function WikiPage() {
  let articles: Article[] = [];
  let error = false;
  try {
    const data = await getArticles();
    articles = data.article || [];
  } catch {
    error = true;
  }

  // グローバルエラー
  if (error) {
    return (
      <div className='max-w-4xl mx-auto p-6 text-center'>
        <h1 className='text-2xl font-bold mb-4'>Wiki</h1>
        <p className='text-gray-500'>
          記事の取得に失敗しました。サービスが起動しているか確認してください。
        </p>
      </div>
    );
  }

  return <WikiClient articles={articles} />;
}
