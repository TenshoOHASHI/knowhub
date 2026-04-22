import Link from 'next/link';
import { getProfile } from '@/lib/api';
import { FiBook, FiBriefcase } from 'react-icons/fi';

export default async function TopPage() {
  let profile = null;
  try {
    const data = await getProfile();
    profile = data.Profile;
  } catch {
    // Todo ログもしくはエラーページに遷移
  }

  return (
    <div className='max-w-4xl mx-auto p-6'>
      <section className='py-12 text-center'>
        <h1 className='text-4xl font-bold mb-4'>
          {profile ? profile.title : 'knowhub'}
        </h1>
        <p className='text-gray-600 dark:text-stone-400 text-lg mb-8'>
          {profile ? profile.bio : '技術ナレッジベースプラットフォーム'}
        </p>
        <div className='flex justify-center gap-4'>
          <Link
            href='/wiki'
            className='flex items-center bg-stone-900 dark:bg-stone-100 text-white dark:text-stone-900 px-6 py-3 rounded-lg hover:bg-stone-800 dark:hover:bg-stone-200'
          >
            <FiBook className='mr-2' />
            Wikiを見る
          </Link>
          <Link
            href='/portfolio'
            className='flex items-center bg-stone-900 dark:bg-stone-100 text-white dark:text-stone-900 px-6 py-3 rounded-lg hover:bg-stone-800 dark:hover:bg-stone-200'
          >
            <FiBriefcase className='mr-2' />
            Portfolio
          </Link>
        </div>
      </section>
    </div>
  );
}
