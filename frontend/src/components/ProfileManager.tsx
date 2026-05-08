'use client';

import { saveProfile as apiSaveProfile } from '@/lib/api';
import { useToast } from '@/context/ToastContext';
import { getProfile } from '@/lib/api';
import { Profile } from '@/lib/types';
import Link from 'next/link';
import { useEffect, useState, useCallback } from 'react';
import ImageUploader from './ImageUploader';
import { FiUser, FiCheck } from 'react-icons/fi';

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
    <div className='p-6 space-y-6'>
      {/* ヘッダー */}
      <div>
        <h2 className='text-lg font-semibold text-stone-900 dark:text-stone-100 mb-1'>プロフィール設定</h2>
        <p className='text-sm text-stone-500 dark:text-stone-400'>あなたのプロフィール情報を管理します</p>
      </div>

      {/* フォーム */}
      <form onSubmit={handleSubmit} className='space-y-4'>
        {profile?.id && <input type='hidden' name='id' value={profile?.id} />}

        <div className='grid grid-cols-1 md:grid-cols-2 gap-4'>
          {/* タイトル */}
          <div>
            <label className='block text-sm font-medium mb-1.5 text-stone-700 dark:text-stone-300'>タイトル</label>
            <input
              type='text'
              name='title'
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder='プロフィールタイトル'
              className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-3 py-2 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-stone-500 transition'
            />
          </div>

          {/* 紹介 */}
          <div className='md:col-span-2'>
            <label className='block text-sm font-medium mb-1.5 text-stone-700 dark:text-stone-300'>自己紹介</label>
            <textarea
              name='bio'
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              placeholder='プロフィール自己紹介'
              rows={3}
              className='w-full border border-stone-300 dark:border-stone-600 rounded-xl px-4 py-2.5 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 transition resize-none'
            />
          </div>

          {/* GitHub URL */}
          <div>
            <label className='block text-sm font-medium mb-1.5 text-stone-700 dark:text-stone-300'>GitHub</label>
            <input
              type='text'
              name='github_url'
              value={githubUrl}
              onChange={(e) => setGithubUrl(e.target.value)}
              placeholder='https://github.com/username'
              className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-3 py-2 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-stone-500 transition'
            />
          </div>

          {/* Twitter URL */}
          <div>
            <label className='block text-sm font-medium mb-1.5 text-stone-700 dark:text-stone-300'>Twitter</label>
            <input
              type='text'
              name='twitter_url'
              value={twitterUrl}
              onChange={(e) => setTwitterUrl(e.target.value)}
              placeholder='https://twitter.com/username'
              className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-3 py-2 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-stone-500 transition'
            />
          </div>

          {/* LinkedIn URL */}
          <div>
            <label className='block text-sm font-medium mb-1.5 text-stone-700 dark:text-stone-300'>LinkedIn</label>
            <input
              type='text'
              name='linkedin_url'
              value={linkedinUrl}
              onChange={(e) => setLinkedinUrl(e.target.value)}
              placeholder='https://linkedin.com/in/username'
              className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-3 py-2 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-stone-500 transition'
            />
          </div>

          {/* Wantedly URL */}
          <div>
            <label className='block text-sm font-medium mb-1.5 text-stone-700 dark:text-stone-300'>Wantedly</label>
            <input
              type='text'
              name='wantedly_url'
              value={wantedlyUrl}
              onChange={(e) => setWantedlyUrl(e.target.value)}
              placeholder='https://www.wantedly.com/id/username'
              className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-3 py-2 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-stone-500 transition'
            />
          </div>

          {/* Avatar URL */}
          <div className='md:col-span-2'>
            <ImageUploader
              value={avatarUrl}
              onChange={(url) => setAvatarUrl(url)}
              label='アバター画像'
              shape='circle'
            />
          </div>

          {/* Skills */}
          <div className='md:col-span-2'>
            <label className='block text-sm font-medium mb-1.5 text-stone-700 dark:text-stone-300'>スキル（カンマ区切り）</label>
            <input
              type='text'
              name='skills'
              value={skillsInput}
              onChange={(e) => setSkillsInput(e.target.value)}
              placeholder='Go, TypeScript, Docker, Next.js'
              className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-3 py-2 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-stone-500 transition'
            />
          </div>

          {/* Languages */}
          <div className='md:col-span-2'>
            <label className='block text-sm font-medium mb-1.5 text-stone-700 dark:text-stone-300'>言語（名前: レベル、カンマ区切り）</label>
            <input
              type='text'
              name='languages'
              value={languagesInput}
              onChange={(e) => setLanguagesInput(e.target.value)}
              placeholder='Japanese: Native, English: Business'
              className='w-full border border-stone-300 dark:border-stone-600 rounded-lg px-3 py-2 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-stone-500 transition'
            />
          </div>
        </div>

        {/* ボタン */}
        <div className='flex items-center gap-3'>
          <button
            type='submit'
            disabled={saving}
            className='bg-stone-900 text-white dark:bg-stone-100 dark:text-stone-900 px-6 py-2.5 rounded-lg text-sm font-medium hover:bg-stone-800 dark:hover:bg-stone-200 transition-colors disabled:opacity-50 disabled:cursor-not-allowed'
          >
            {saving ? '保存中...' : '更新'}
          </button>

          {message && message.includes('プロフィールを更新しました') && (
            <Link
              href='/profile'
              className='text-sm text-stone-600 dark:text-stone-400 underline hover:text-stone-900 dark:hover:text-stone-200 flex items-center gap-1'
            >
              <FiCheck size={14} />
              Profileで確認 →
            </Link>
          )}
        </div>
      </form>
    </div>
  );
}
