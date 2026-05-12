import type { ReactNode } from 'react';

export default function StickyNote({ children }: { children: ReactNode }) {
  return (
    <div className='sticky-note'>
      {/* ピン */}
      <div className='sticky-note-pin'>
        <svg width='16' height='16' viewBox='0 0 24 24' fill='none'>
          <circle cx='12' cy='8' r='5' fill='#ef4444' stroke='#b91c1c' strokeWidth='1' />
          <circle cx='12' cy='8' r='2' fill='#fca5a5' opacity='0.6' />
          <line x1='12' y1='13' x2='12' y2='20' stroke='#991b1b' strokeWidth='1.5' />
        </svg>
      </div>
      <div className='sticky-note-body text-sm'>
        {children}
      </div>
    </div>
  );
}
