'use client';

import { saveProfile as apiSaveProfile } from '@/lib/api';
import { useToast } from '@/context/ToastContext';
import { getProfile } from '@/lib/api';
import { Profile } from '@/lib/types';
import Link from 'next/link';
import { useEffect, useState, useCallback } from 'react';
import ImageUploader from './ImageUploader';

export default function ProfileManager() {
  const [profile, setProfile] = useState<Profile | null>(null);
  const [title, setTitle] = useState('');
  const [bio, setBio] = useState('');
  const [githubUrl, setGithubUrl] = useState('');
  const [avatarUrl, setAvatarUrl] = useState('');
  const [twitterUrl, setTwitterUrl] = useState('');
  const [linkedinUrl, setLinkedinUrl] = useState('');
  const [wantedlyUrl, setWantedlyUrl] = useState('');
  const [skillsInput, setSkillsInput] = useState('');
  const [languagesInput, setLanguagesInput] = useState('');
  const [message, setMessage] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const { toast } = useToast();

  const load = useCallback(() => {
    getProfile()
      .then((data) => {
        const p = data.Profile || {};
        setProfile(p);
        setTitle(p.title || '');
        setBio(p.bio || '');
        setGithubUrl(p.github_url || '');
        setAvatarUrl(p.avatar_url || '');
        setTwitterUrl(p.twitter_url || '');
        setLinkedinUrl(p.linkedin_url || '');
        setWantedlyUrl(p.wantedly_url || '');
        try {
          const parsed = JSON.parse(p.skills || '[]');
          setSkillsInput(Array.isArray(parsed) ? parsed.join(', ') : '');
        } catch {
          setSkillsInput('');
        }
        try {
          const parsed = JSON.parse(p.languages || '[]');
          if (Array.isArray(parsed)) {
            setLanguagesInput(
              parsed
                .map(
                  (l: { name: string; level: string }) =>
                    `${l.name}: ${l.level}`,
                )
                .join(', '),
            );
          }
        } catch {
          setLanguagesInput('');
        }
      })
      .catch(() => toast('取得に失敗しました', 'error'));
  }, [toast]);

  useEffect(() => {
    load();
  }, [load]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (saving) return;
    setSaving(true);
    try {
      const skills = JSON.stringify(
        skillsInput
          .split(',')
          .map((s) => s.trim())
          .filter(Boolean),
      );
      const languages = JSON.stringify(
        languagesInput
          .split(',')
          .map((s) => s.trim())
          .filter(Boolean)
          .map((s) => {
            const [name, ...rest] = s.split(':');
            return { name: name.trim(), level: rest.join(':').trim() };
          })
          .filter((l) => l.name && l.level),
      );

      await apiSaveProfile({
        title,
        bio,
        github_url: githubUrl,
        avatar_url: avatarUrl,
        twitter_url: twitterUrl,
        linkedin_url: linkedinUrl,
        wantedly_url: wantedlyUrl,
        skills,
        languages,
      });
      toast('プロフィールを更新しました', 'success');
      setMessage('プロフィールを更新しました');
    } catch {
      toast('更新に失敗しました', 'error');
      setMessage('更新に失敗しました');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className='max-w-4xl mx-auto p-4'>
      <h1 className='text-3xl mb-4 font-bold'>プロファイル</h1>

      {/* フォーム */}
      <form onSubmit={handleSubmit} className='space-y-4'>
        {profile?.id && <input type='hidden' name='id' value={profile?.id} />}

        {/* タイトル */}
        <div>
          <label className='block mb-1'>タイトル</label>
          <input
            type='text'
            name='title'
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder='プロフィールタイトル'
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 mb-2 bg-transparent'
          />
        </div>

        {/* 紹介 */}
        <div>
          <label className='block mb-1'>紹介</label>
          <textarea
            name='bio'
            value={bio}
            onChange={(e) => setBio(e.target.value)}
            placeholder='プロフィール自己紹介'
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 mb-2 bg-transparent'
          />
        </div>

        {/* GitHub URL */}
        <div>
          <label className='block mb-1'>GitHub URL</label>
          <input
            type='text'
            name='github_url'
            value={githubUrl}
            onChange={(e) => setGithubUrl(e.target.value)}
            placeholder='https://github.com/username'
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 mb-2 bg-transparent'
          />
        </div>

        {/* Avatar URL */}
        <ImageUploader
          value={avatarUrl}
          onChange={(url) => setAvatarUrl(url)}
          label='Avatar'
          shape='circle'
        />

        {/* Twitter URL */}
        <div>
          <label className='block mb-1'>Twitter URL</label>
          <input
            type='text'
            name='twitter_url'
            value={twitterUrl}
            onChange={(e) => setTwitterUrl(e.target.value)}
            placeholder='https://twitter.com/username'
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 mb-2 bg-transparent'
          />
        </div>

        {/* LinkedIn URL */}
        <div>
          <label className='block mb-1'>LinkedIn URL</label>
          <input
            type='text'
            name='linkedin_url'
            value={linkedinUrl}
            onChange={(e) => setLinkedinUrl(e.target.value)}
            placeholder='https://linkedin.com/in/username'
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 mb-2 bg-transparent'
          />
        </div>

        {/* Wantedly URL */}
        <div>
          <label className='block mb-1'>Wantedly URL</label>
          <input
            type='text'
            name='wantedly_url'
            value={wantedlyUrl}
            onChange={(e) => setWantedlyUrl(e.target.value)}
            placeholder='https://www.wantedly.com/id/username'
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 mb-2 bg-transparent'
          />
        </div>

        {/* Skills */}
        <div>
          <label className='block mb-1'>Skills（カンマ区切り）</label>
          <input
            type='text'
            name='skills'
            value={skillsInput}
            onChange={(e) => setSkillsInput(e.target.value)}
            placeholder='Go, TypeScript, Docker, Next.js'
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 mb-2 bg-transparent'
          />
        </div>

        {/* Languages */}
        <div>
          <label className='block mb-1'>
            Languages（名前: レベル, カンマ区切り）
          </label>
          <input
            type='text'
            name='languages'
            value={languagesInput}
            onChange={(e) => setLanguagesInput(e.target.value)}
            placeholder='Japanese: Native, English: Business'
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 mb-2 bg-transparent'
          />
        </div>

        <button
          type='submit'
          disabled={saving}
          className='px-4 py-2 rounded-lg bg-green-700 dark:bg-green-300 text-white dark:text-stone-900 hover:bg-green-800 dark:hover:bg-green-300/90 mb-2'
        >
          {saving ? '保存中...' : '更新'}
        </button>
      </form>

      {message && message.includes('プロフィールを更新しました') && (
        <div>
          <Link
            href='/profile'
            className='underline hover:text-green-800 dark:hover:text-green-300'
          >
            Profileで確認 →
          </Link>
        </div>
      )}
    </div>
  );
}
