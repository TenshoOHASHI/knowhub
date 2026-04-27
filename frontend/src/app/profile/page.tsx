'use client';

import { getProfile } from '@/lib/api';
import { motion } from 'motion/react';
import { useEffect, useState } from 'react';
import { FiGithub, FiTwitter, FiLinkedin } from 'react-icons/fi';
import { MdWork } from 'react-icons/md';

import {
  SiGo,
  SiTypescript,
  SiDocker,
  SiNextdotjs,
  SiReact,
  SiGrafana,
  SiMysql,
  SiRedis,
  SiKubernetes,
  SiTerraform,
  SiGin,
  SiRust,
  SiPython,
  SiPostgresql,
  SiLinux,
  SiGit,
  SiSupabase,
  SiClaude,
  SiWantedly,
} from 'react-icons/si';
import type { Profile } from '@/lib/types';
import Image from 'next/image';

// Protocol Buffers custom SVG icon (not in react-icons)
const SiProtobuf: React.FC<{ className?: string }> = ({ className }) => (
  <svg
    className={className}
    viewBox='0 0 24 24'
    fill='currentColor'
    xmlns='http://www.w3.org/2000/svg'
  >
    <path d='M12 0L1.608 6v12L12 24l10.392-6V6L12 0zm0 2.16l8.392 4.85v9.98L12 21.84l-8.392-4.85V7.01L12 2.16zM12 6L7.5 8.625v6.75L12 18l4.5-2.625v-6.75L12 6zm0 2.571l2.25 1.304v2.25L12 13.429l-2.25-1.304v-2.25L12 8.571z' />
  </svg>
);

const iconMap: Record<string, React.ComponentType<{ className?: string }>> = {
  Go: SiGo,
  TypeScript: SiTypescript,
  Docker: SiDocker,
  'Next.js': SiNextdotjs,
  React: SiReact,
  Grafana: SiGrafana,
  MySQL: SiMysql,
  Redis: SiRedis,
  Kubernetes: SiKubernetes,
  Terraform: SiTerraform,
  Gin: SiGin,
  Rust: SiRust,
  Python: SiPython,
  PostgreSQL: SiPostgresql,
  Linux: SiLinux,
  Git: SiGit,
  Supabase: SiSupabase,
  Claude: SiClaude,
  ProtocolBuffers: SiProtobuf,
  Protobuf: SiProtobuf,
};

interface Language {
  name: string;
  level: string;
}

export default function ProfilePage() {
  const [profile, setProfile] = useState<Profile | null>(null);
  const [skills, setSkills] = useState<string[]>([]);
  const [languages, setLanguages] = useState<Language[]>([]);

  useEffect(() => {
    getProfile()
      .then((data) => {
        const p = data.Profile;
        setProfile(p);
        try {
          const parsed = JSON.parse(p.skills || '[]');
          if (Array.isArray(parsed)) setSkills(parsed);
        } catch {
          setSkills([]);
        }
        try {
          const parsed = JSON.parse(p.languages || '[]');
          if (Array.isArray(parsed)) setLanguages(parsed);
        } catch {
          setLanguages([]);
        }
      })
      .catch(() => {});
  }, []);

  if (!profile) {
    return (
      <div className='flex items-center justify-center min-h-[60vh]'>
        <div className='animate-pulse text-gray-400'>Loading...</div>
      </div>
    );
  }

  return (
    <div className='max-w-4xl mx-auto px-6 py-12 space-y-16'>
      {/* Hero Section */}
      <motion.section
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6 }}
        className='flex flex-col md:flex-row items-center gap-8'
      >
        {/* Avatar */}
        <div className='shrink-0'>
          {profile.avatar_url ? (
            <Image
              width={256}
              height={256}
              src={profile.avatar_url.trim()}
              alt={profile.title}
              className='w-48 h-48 rounded-full object-cover border-2 border-stone-300 dark:border-stone-600 shadow-lg'
            />
          ) : (
            <div className='w-32 h-32 rounded-full bg-stone-200 dark:bg-stone-700 flex items-center justify-center text-4xl font-bold text-stone-500 dark:text-stone-400'>
              {profile.title.charAt(0).toUpperCase()}
            </div>
          )}
        </div>

        {/* Info */}
        <div className='flex-1 text-center md:text-left'>
          <h1 className='text-3xl md:text-4xl font-bold mb-2'>
            {profile.title}
          </h1>
          <p className='text-gray-600 dark:text-stone-400 mb-4 text-lg'>
            {profile.bio}
          </p>
          <span className='inline-flex items-center gap-1.5 text-sm text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900/20 px-3 py-1 rounded-full'>
            <MdWork />
            Available for work
          </span>

          {/* Social Links */}
          <div className='flex items-center gap-4 mt-4 justify-center md:justify-start'>
            {profile.github_url && (
              <a
                href={profile.github_url}
                target='_blank'
                rel='noopener noreferrer'
                className='text-gray-500 hover:text-black dark:hover:text-white transition-colors'
              >
                <FiGithub size={22} />
              </a>
            )}
            {profile.twitter_url && (
              <a
                href={profile.twitter_url}
                target='_blank'
                rel='noopener noreferrer'
                className='text-gray-500 hover:text-black dark:hover:text-white transition-colors'
              >
                <FiTwitter size={22} />
              </a>
            )}
            {profile.linkedin_url && (
              <a
                href={profile.linkedin_url}
                target='_blank'
                rel='noopener noreferrer'
                className='text-gray-500 hover:text-black dark:hover:text-white transition-colors'
              >
                <FiLinkedin size={22} />
              </a>
            )}
            {profile.wantedly_url && (
              <a
                href={profile.wantedly_url}
                target='_blank'
                rel='noopener noreferrer'
                className='text-gray-500 hover:text-black dark:hover:text-white transition-colors'
                title='Wantedly'
              >
                <SiWantedly size={22} />
              </a>
            )}
          </div>
        </div>
      </motion.section>

      {/* Skills Section */}
      {skills.length > 0 && (
        <motion.section
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.4 }}
        >
          <h2 className='text-2xl font-bold mb-6'>Skills</h2>
          <div className='flex flex-wrap gap-3'>
            {skills.map((skill, index) => {
              const Icon = iconMap[skill];
              return (
                <motion.div
                  key={skill}
                  initial={{ opacity: 0, y: 10 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true }}
                  transition={{ delay: index * 0.08, duration: 0.3 }}
                  whileHover={{ scale: 1.08 }}
                  className='flex items-center gap-2 px-4 py-2 rounded-lg bg-stone-100 dark:bg-stone-800 text-sm font-medium border border-stone-200 dark:border-stone-700'
                >
                  {Icon && <Icon className='text-lg' />}
                  {skill}
                </motion.div>
              );
            })}
          </div>
        </motion.section>
      )}

      {/* Languages Section */}
      {languages.length > 0 && (
        <motion.section
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.4 }}
        >
          <h2 className='text-2xl font-bold mb-6'>Languages</h2>
          <div className='space-y-3'>
            {languages.map((lang, index) => (
              <motion.div
                key={lang.name}
                initial={{ opacity: 0, x: -10 }}
                whileInView={{ opacity: 1, x: 0 }}
                viewport={{ once: true }}
                transition={{ delay: index * 0.1, duration: 0.3 }}
                className='flex items-center justify-between gap-4 max-w-md'
              >
                <span className='text-sm font-medium w-24'>{lang.name}</span>
                <div className='flex-1 h-2 bg-stone-200 dark:bg-stone-700 rounded-full overflow-hidden'>
                  <motion.div
                    initial={{ width: 0 }}
                    whileInView={{
                      width:
                        lang.level === 'Native'
                          ? '100%'
                          : lang.level === 'Fluent'
                            ? '80%'
                            : lang.level === 'Business'
                              ? '60%'
                              : lang.level === 'Conversational'
                                ? '40%'
                                : '50%',
                    }}
                    viewport={{ once: true }}
                    transition={{ delay: index * 0.1 + 0.3, duration: 0.6 }}
                    className='h-full bg-green-500 dark:bg-green-400 rounded-full'
                  />
                </div>
                <span className='text-xs text-gray-500 dark:text-stone-400 w-28 text-right'>
                  {lang.level}
                </span>
              </motion.div>
            ))}
          </div>
        </motion.section>
      )}
    </div>
  );
}
