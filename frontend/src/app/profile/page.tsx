import { getProfile } from '@/lib/api';
import { FiGithub } from 'react-icons/fi';

export default async function ProfilePage() {
  const data = await getProfile();
  const profile = data.Profile;

  return (
    <div className='max-w-4xl mx-auto p-6'>
      <h1 className='text-3xl font-bold mb-2'>{profile.title}</h1>
      <p className='text-gray-600 dark:text-stone-400 mb-4'>{profile.bio}</p>
      <a
        href={profile.github_url}
        target='_blank'
        className='inline-flex items-center hover:underline'
      >
        <FiGithub className='mr-1' />
        GitHub
      </a>
    </div>
  );
}
