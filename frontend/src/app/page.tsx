'use client';

import Link from 'next/link';
import { motion } from 'motion/react';
import {
  FiBookOpen,
  FiBriefcase,
  FiUser,
  FiArrowRight,
  FiZap,
  FiCode,
  FiEdit3,
  FiCpu,
  FiGlobe,
  FiSun,
} from 'react-icons/fi';

const keywords = [
  'Go',
  'gRPC',
  'CQRS',
  'Next.js',
  'TypeScript',
  'Docker',
  'MySQL',
  'Redis',
  'Protocol Buffers',
  'Microservices',
  'React',
  'JavaScript',
  'Tailwind',
  'Git',
];

const values = [
  { icon: FiSun, label: '作る' },
  { icon: FiCpu, label: 'AI共存' },
  { icon: FiCode, label: '探究' },
  { icon: FiEdit3, label: '学ぶ' },
  { icon: FiZap, label: '加速' },
  { icon: FiGlobe, label: '進歩' },
];

export default function TopPage() {
  return (
    <div className='h-full overflow-y-auto flex flex-col items-center justify-start pt-[8vh] px-6 pb-6'>
      {/* Floating keywords */}
      <div className='fixed inset-0 overflow-hidden pointer-events-none select-none'>
        {keywords.map((kw, i) => {
          // 内側に寄せる（中央付近に集める）
          const top = 10 + ((i * 31) % 50); // 20%〜70%
          const left = 25 + ((i * 19) % 55); // 20%〜75%

          // 外側に発散させる
          // const top = 3 + ((i * 29) % 90); // 上端3%〜下端93%
          // const left = 2 + ((i * 23) % 92); // 左端2%〜右端94%

          const delay = i * 0.5;
          const size = 0.65 + (i % 3) * 0.1;
          return (
            <motion.span
              key={kw}
              className='absolute text-stone-200 dark:text-stone-700 font-mono font-semibold whitespace-nowrap'
              style={{
                top: `${top}%`,
                left: `${left}%`,
                fontSize: `${size}rem`,
              }}
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay, duration: 1, ease: 'easeOut' }}
            >
              {kw}
            </motion.span>
          );
        })}
      </div>

      {/* Title */}
      <motion.div
        className='relative flex items-center gap-4'
        initial={{ opacity: 0, y: 30 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.7 }}
      >
        <h1 className='text-5xl md:text-7xl font-bold tracking-tight'>
          Ten
          <span className='bg-gradient-to-r from-stone-700 to-stone-400 bg-clip-text text-transparent'>
            Hub
          </span>
        </h1>
      </motion.div>

      {/* Tagline */}
      <motion.p
        className='relative mt-4 text-base md:text-lg text-gray-500 dark:text-stone-400 max-w-md leading-relaxed text-center'
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.3, duration: 0.6 }}
      >
        技術を記し、考え、発信する。
        <br />
        小さなアウトプットが、やがてプロダクトになる。
      </motion.p>

      {/* Typing line */}
      <motion.div
        className='relative mt-3 font-mono text-xs text-stone-400 dark:text-stone-500'
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.8, duration: 0.5 }}
      >
        <span className='border-r-2 border-stone-400 dark:border-stone-500 pr-1 animate-pulse'>
          build &gt; learn &gt; share &gt; repeat_
        </span>
      </motion.div>

      {/* CTA buttons */}
      <motion.div
        className='relative flex flex-wrap justify-center gap-3 mt-4 mb-18'
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.6, duration: 0.5 }}
      >
        <Link
          href='/wiki'
          className='group flex items-center gap-2 bg-stone-900 dark:bg-stone-100 text-white dark:text-stone-900 px-5 py-2.5 rounded-lg hover:bg-stone-800 dark:hover:bg-stone-200 transition-colors text-sm'
        >
          <FiBookOpen />
          Wiki
          <FiArrowRight className='group-hover:translate-x-1 transition-transform' />
        </Link>
        <Link
          href='/portfolio'
          className='group flex items-center gap-2 border border-stone-300 dark:border-stone-600 px-5 py-2.5 rounded-lg hover:border-stone-500 dark:hover:border-stone-400 transition-colors text-sm'
        >
          <FiBriefcase />
          Portfolio
          <FiArrowRight className='group-hover:translate-x-1 transition-transform' />
        </Link>
        <Link
          href='/profile'
          className='group flex items-center gap-2 border border-stone-300 dark:border-stone-600 px-5 py-2.5 rounded-lg hover:border-stone-500 dark:hover:border-stone-400 transition-colors text-sm'
        >
          <FiUser />
          Profile
        </Link>
      </motion.div>

      {/* Spacer */}
      <div className='h-10' />

      {/* Mind Map */}
      <motion.div
        className='relative max-w-lg w-full mb-6'
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.9, duration: 0.5 }}
      >
        {/* Center node */}
        <div className='flex justify-center mb-4'>
          <motion.div
            className='px-6 py-3 rounded-full bg-stone-900 dark:bg-stone-100 text-white dark:text-stone-900 text-sm font-bold'
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ delay: 0.9, duration: 0.4, type: 'spring' }}
          >
            Mindset
          </motion.div>
        </div>

        {/* Branch lines + nodes */}
        <div className='grid grid-cols-3 gap-x-6 gap-y-3'>
          {values.map((v, i) => {
            const Icon = v.icon;
            return (
              <motion.div
                key={v.label}
                className='flex flex-col items-center'
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 1 + i * 0.1, duration: 0.3 }}
              >
                {/* Connector line */}
                <div className='w-px h-3 bg-stone-300 dark:bg-stone-600' />
                {/* Node */}
                <div className='flex items-center gap-2 px-3 py-2 rounded-full bg-stone-50 dark:bg-stone-800/50 border border-stone-200 dark:border-stone-700/50'>
                  <Icon className='text-sm shrink-0' />
                  <span className='text-xs font-medium whitespace-nowrap'>
                    {v.label}
                  </span>
                </div>
              </motion.div>
            );
          })}
        </div>
      </motion.div>

      {/* What is knowhub? */}
      <motion.div
        className='relative mt-8 text-center max-w-xl'
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 1.4, duration: 0.4 }}
      >
        <h2 className='text-3xl font-semibold dark:text-stone-300 leading-relaxed'>
          <span className='text-4xl'>W</span>hat is Tenhub
        </h2>
        <p className='text-gray-500 dark:text-stone-400 text-md leading-relaxed relative group cursor-default'>
          <span className='relative inline-block'>
            学びの断片を、知識の資産へ変えるナレッジベース。
            <span className='absolute left-0 bottom-0 w-full h-0.5 bg-red-500/60 origin-left scale-x-0 group-hover:scale-x-100 transition-transform duration-500' />
          </span>
          <br />
          <span className='relative inline-block'>
            Go マイクロサービス + Next.js
            で構築された、モダンな技術スタックの実践場。
            <span className='absolute left-0 bottom-0 w-full h-0.5 bg-red-500/60 origin-left scale-x-0 group-hover:scale-x-100 transition-transform duration-500 delay-500' />
          </span>
        </p>
      </motion.div>
    </div>
  );
}
