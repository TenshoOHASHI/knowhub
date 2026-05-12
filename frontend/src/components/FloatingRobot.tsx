'use client';

import { useEffect, useState, useRef } from 'react';

const MESSAGES = [
  '生成中だよ、まってね！',
  'もう少しで完成するよ！',
  '考え中...',
  'ローカルのOllamaは少し生成が遅いかも',
  '1日の実行回数は6回までだよ',
  'Wikiの知識を探しています...',
  'もうちょっとだけ待ってね！',
  '一生懸命考えてるよ！',
];

function randomMessage() {
  return MESSAGES[Math.floor(Math.random() * MESSAGES.length)];
}

export default function FloatingRobot({ visible }: { visible: boolean }) {
  const [currentMessage, setCurrentMessage] = useState(randomMessage);
  const [position, setPosition] = useState(20);
  const [direction, setDirection] = useState<'down' | 'up'>('down');
  const moveRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const messageRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // 縦方向の歩行移動
  useEffect(() => {
    if (!visible) return;

    let pos = 20;
    let dir: 'down' | 'up' = 'down';

    moveRef.current = setInterval(() => {
      if (dir === 'down') {
        pos += 0.4;
        if (pos >= 75) dir = 'up';
      } else {
        pos -= 0.4;
        if (pos <= 15) dir = 'down';
      }
      setPosition(pos);
      setDirection(dir);
    }, 60);

    return () => {
      if (moveRef.current) clearInterval(moveRef.current);
      setPosition(20);
      setDirection('down');
    };
  }, [visible]);

  // メッセージ切替
  useEffect(() => {
    if (!visible) return;
    messageRef.current = setInterval(() => {
      setCurrentMessage(randomMessage());
    }, 3500);
    return () => {
      if (messageRef.current) clearInterval(messageRef.current);
      setCurrentMessage(randomMessage());
    };
  }, [visible]);

  if (!visible) return null;

  const isUp = direction === 'up';

  return (
    <div
      className='absolute right-2 z-10 pointer-events-none flex items-center gap-1'
      style={{ top: `${position}%`, transform: 'translateY(-50%)' }}
    >
      {/* 吹き出し（ロボットの左側） */}
      <div className='bubble-fade relative bg-white dark:bg-stone-800 px-2.5 py-1.5 rounded-xl shadow-md border border-stone-200 dark:border-stone-600 text-[11px] text-stone-600 dark:text-stone-300 whitespace-nowrap'>
        {currentMessage}
        {/* 右向き三角 */}
        <div className='absolute top-1/2 -right-1.5 -translate-y-1/2 w-0 h-0 border-t-[5px] border-t-transparent border-b-[5px] border-b-transparent border-l-[6px] border-l-white dark:border-l-stone-800' />
      </div>
      {/* ロボット */}
      <div className='walking-robot-body shrink-0'>
        <svg width='40' height='60' viewBox='0 0 120 180' className='text-stone-400 dark:text-stone-500 overflow-visible'>
          {/* 頭 */}
          <rect x='28' y='10' width='64' height='30' rx='10' fill='currentColor' />
          {/* 顔スクリーン */}
          <rect x='33' y='16' width='54' height='20' rx='5' fill='#1e293b' opacity='0.3' />
          {/* アンテナ */}
          <line x1='60' y1='10' x2='60' y2='2' stroke='currentColor' strokeWidth='2.5' />
          <circle cx='60' cy='0' r='3' fill='#60a5fa' />
          {/* 目 */}
          <circle cx='45' cy='26' r='5' fill='#3b82f6' />
          <circle cx='75' cy='26' r='5' fill='#3b82f6' />
          {/* 口 */}
          <rect x='52' y='32' width='16' height='2' rx='1' fill='white' opacity='0.7' />
          {/* 体 */}
          <rect x='35' y='42' width='50' height='28' rx='8' fill='currentColor' />
          {/* Tマーク */}
          <g fill='#1e40af'>
            <rect x='58.5' y='50' width='3' height='12' rx='0.5' />
            <rect x='51' y='50' width='18' height='3' rx='0.5' />
          </g>
          {/* 左腕 */}
          <rect x='20' y='46' width='14' height='5' rx='2.5' fill='currentColor' />
          <circle cx='18' cy='48' r='5' fill='currentColor' />
          {/* 右腕 */}
          <rect x='86' y='46' width='14' height='5' rx='2.5' fill='currentColor' />
          <circle cx='102' cy='48' r='5' fill='currentColor' />
          {/* 左足 */}
          <g className='walking-robot-leg-left'>
            <rect x='40' y='70' width='12' height='18' rx='4' fill='currentColor' />
            <rect x='36' y='86' width='18' height='10' rx='5' fill='currentColor' />
          </g>
          {/* 右足 */}
          <g className='walking-robot-leg-right'>
            <rect x='68' y='70' width='12' height='18' rx='4' fill='currentColor' />
            <rect x='64' y='86' width='18' height='10' rx='5' fill='currentColor' />
          </g>

          {/* ブーストエフェクト（足元） */}
          <g className={isUp ? 'robot-boost-strong' : 'robot-boost-weak'}>
            {/* グロー（上昇時に大きく光る） */}
            <ellipse cx='60' cy='105' rx={isUp ? 22 : 8} ry={isUp ? 22 : 8} fill='#f97316' opacity={isUp ? 0.15 : 0.05} />

            {/* 外側の炎 */}
            <ellipse cx='60' cy='108' rx={isUp ? 16 : 5} ry={isUp ? 30 : 8} fill='#f97316' opacity={isUp ? 0.7 : 0.2} />
            {/* 中間の炎 */}
            <ellipse cx='60' cy='106' rx={isUp ? 10 : 3} ry={isUp ? 22 : 5} fill='#fbbf24' opacity={isUp ? 0.85 : 0.25} />
            {/* 芯の炎 */}
            <ellipse cx='60' cy='104' rx={isUp ? 5 : 1.5} ry={isUp ? 14 : 3} fill='#fef3c7' opacity={isUp ? 1 : 0.3} />

            {/* パーティクル（上昇時に大きく飛び散る） */}
            <circle className='boost-particle-1' cx='48' cy='120' r={isUp ? 4 : 1.5} fill='#fb923c' opacity={isUp ? 0.8 : 0.1} />
            <circle className='boost-particle-2' cx='60' cy='130' r={isUp ? 3.5 : 1} fill='#fbbf24' opacity={isUp ? 0.7 : 0.08} />
            <circle className='boost-particle-3' cx='72' cy='124' r={isUp ? 4 : 1.5} fill='#fb923c' opacity={isUp ? 0.8 : 0.1} />
          </g>
        </svg>
      </div>
    </div>
  );
}
