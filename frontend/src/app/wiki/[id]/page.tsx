import { getArticle } from '@/lib/api';
import Link from 'next/link';
import { FiArrowLeft, FiLock, FiUnlock } from 'react-icons/fi';
import { MdOutlineSupportAgent } from 'react-icons/md';
import ArticleContent from '@/components/ArticleContent';
import ArticleActions from '@/components/ArticleActions';
import { extractToc } from '@/lib/toc';
import { TableOfContents } from '@/components/TableOfContents';
import { cookies } from 'next/headers';
import type { Metadata } from 'next';

interface Props {
  params: Promise<{ id: string }>;
}

// 動的メタデータ生成
export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params;

  let article = null;
  try {
    const data = await getArticle(id, undefined); // 公開記事のみ取得
    article = data.Article;
  } catch {
    // 記事が見つからない場合
  }

  if (!article) {
    return {
      title: '記事が見つかりません - TenHub',
    };
  }

  // 記事の内容から説明文を作成（最初の100文字程度）
  const plainText = article.content.replace(/[#*`>\-\[\]()]/g, '').trim();
  const description = plainText.slice(0, 100) + (plainText.length > 100 ? '...' : '');

  return {
    title: `${article.title} - TenHub Wiki`,
    description: description || '学びの断片を、知識の資産へ変えるナレッジベース',
    openGraph: {
      title: `${article.title} - TenHub Wiki`,
      description: description || '学びの断片を、知識の資産へ変えるナレッジベース',
      type: 'article',
      url: `https://tenhub.tech/wiki/${id}`,
      siteName: 'TenHub',
      images: ['/api/og'],
      locale: 'ja_JP',
    },
    twitter: {
      card: 'summary_large_image',
      title: `${article.title} - TenHub Wiki`,
      description: description || '学びの断片を、知識の資産へ変えるナレッジベース',
      images: ['/api/og'],
    },
  };
}

export default async function ArticleDetailPage({ params }: Props) {
  const { id } = await params;

  // ブラウザから届いたCookieからtokenを取り出し、Bearer で Gateway に転送
  // Node.js fetch は Cookie ヘッダーを禁止ヘッダーとして除去するため Bearer を使う
  const cookieStore = await cookies();
  const tokenCookie = cookieStore.get('token');
  const token = tokenCookie?.value;

  let article = null;
  try {
    const data = await getArticle(id, token);
    article = data.Article;
  } catch {
    // 403（非公開）やその他エラー
  }

  // 記事が取得できない（非公開 or 存在しない）
  if (!article) {
    return (
      <div className='max-w-4xl mx-auto px-6 py-12 text-center'>
        <div className='inline-flex items-center justify-center w-16 h-16 rounded-full bg-stone-100 dark:bg-stone-800 mb-4'>
          <FiLock size={24} className='text-stone-400' />
        </div>
        <h1 className='text-2xl font-bold mb-2'>この記事は非公開です</h1>
        <p className='text-stone-400 mb-6'>閲覧できません</p>
        <Link
          href='/wiki'
          className='inline-flex items-center gap-1 text-sm text-stone-500 hover:text-black dark:hover:text-stone-100'
        >
          <FiArrowLeft size={14} />
          Wikiに戻る
        </Link>
      </div>
    );
  }
  // {id:string, text:string, level: 2 | 3 }
  const toc = extractToc(article.content);

  return (
    <div className='max-w-6xl mx-auto px-6 py-6'>
      {/* Navigation */}
      <div className='flex flex-col gap-1 mb-6'>
        <Link
          href='/wiki'
          className='inline-flex items-center text-md text-gray-600 dark:text-stone-400 hover:text-black dark:hover:text-stone-100 w-fit'
        >
          <FiArrowLeft className='mr-1' />
          Wikiに戻る
        </Link>
        <div className='flex gap-4'>
          <Link
            href={`/admin?id=${article.id}`}
            className='text-md text-gray-600 hover:text-black dark:text-stone-400 dark:hover:text-stone-100'
          >
            編集
          </Link>
          <Link
            href='/chat'
            className='inline-flex items-center gap-1 text-md text-gray-600 hover:text-black dark:text-stone-400 dark:hover:text-stone-100'
          >
            <MdOutlineSupportAgent size={16} />
            Chat
          </Link>
        </div>
      </div>

      <div className='flex items-start justify-between gap-4 mb-2'>
        <h1 className='text-3xl font-bold flex items-center gap-2 overflow-hidden'>
          {article.visibility === 'locked' && (
            <span className='inline-flex items-center justify-center w-7 h-7 rounded-lg bg-stone-200 dark:bg-stone-700 shrink-0'>
              <FiUnlock
                size={14}
                className='text-stone-500 dark:text-stone-400'
              />
            </span>
          )}
          <span className='min-w-0 break-all text-4xl'>{article.title}</span>
        </h1>
        <ArticleActions articleId={article.id} />
      </div>
      <p className='text-stone-400 mb-6 text-lg'>
        {new Date(article.created_at.seconds * 1000).toLocaleDateString(
          'ja-JP',
        )}
      </p>

      <div className='flex gap-8'>
        {/* Article */}
        <div className='flex-1 min-w-0'>
          <ArticleContent content={article.content} />
        </div>

        {/* TOC Sidebar */}
        {toc.length > 0 && (
          <aside className='hidden lg:block w-56 shrink-0'>
            <TableOfContents items={toc} />
          </aside>
        )}
      </div>
    </div>
  );
}
