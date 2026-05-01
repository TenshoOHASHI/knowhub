import React from 'react';
import type { CalloutType } from '@/lib/remark-callout';

interface CalloutProps {
  type: CalloutType;
  children: React.ReactNode;
}

/* ── SVG icons (16×16) ─────────────────────────────────── */

function IconInfo() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="8" cy="8" r="7" />
      <line x1="8" y1="11" x2="8" y2="7.5" />
      <circle cx="8" cy="5" r="0.5" fill="currentColor" stroke="none" />
    </svg>
  );
}

function IconLightbulb() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round">
      <path d="M6 12.5h4" />
      <path d="M6.5 14h3" />
      <path d="M8 1.5a4 4 0 0 0-2.5 7.1V11h5V8.6A4 4 0 0 0 8 1.5z" />
    </svg>
  );
}

function IconAlertTriangle() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round">
      <path d="M8 2L1.5 13.5h13L8 2z" />
      <line x1="8" y1="6.5" x2="8" y2="9.5" />
      <circle cx="8" cy="11.5" r="0.5" fill="currentColor" stroke="none" />
    </svg>
  );
}

function IconStop() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M4.7 1.5h6.6l3.2 3.2v6.6l-3.2 3.2H4.7l-3.2-3.2V4.7z" />
      <line x1="6" y1="6" x2="10" y2="10" />
      <line x1="10" y1="6" x2="6" y2="10" />
    </svg>
  );
}

function IconFlame() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round">
      <path d="M8 1.5c0 3-3.5 4-3.5 7.5a3.5 3.5 0 0 0 7 0C11.5 5.5 8 4.5 8 1.5z" />
      <path d="M8 10a1.5 1.5 0 0 1-1.5-1.5c0-1.2 1.5-2 1.5-3.5 0 1.5 1.5 2.3 1.5 3.5A1.5 1.5 0 0 1 8 10z" />
    </svg>
  );
}

function IconBroadcast() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="8" cy="10" r="1.5" />
      <path d="M5.3 7.3a3.8 3.8 0 0 1 5.4 0" />
      <path d="M3 5a7 7 0 0 1 10 0" />
    </svg>
  );
}

function IconNote() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round">
      <rect x="2" y="2" width="12" height="12" rx="2" />
      <line x1="5" y1="5.5" x2="11" y2="5.5" />
      <line x1="5" y1="8" x2="11" y2="8" />
      <line x1="5" y1="10.5" x2="8" y2="10.5" />
    </svg>
  );
}

/* ── Theme config ──────────────────────────────────────── */

type ThemeConfig = {
  icon: React.ReactNode;
  label: string;
  border: string;
  bg: string;
  headerText: string;
};

const CONFIG: Record<CalloutType, ThemeConfig> = {
  note: {
    icon: <IconNote />,
    label: 'Note',
    border: 'border-sky-500/60 dark:border-sky-400/50',
    bg: 'bg-sky-50 dark:bg-sky-950/30',
    headerText: 'text-sky-800 dark:text-sky-300',
  },
  info: {
    icon: <IconInfo />,
    label: 'Info',
    border: 'border-blue-500/60 dark:border-blue-400/50',
    bg: 'bg-blue-50 dark:bg-blue-950/30',
    headerText: 'text-blue-800 dark:text-blue-300',
  },
  tip: {
    icon: <IconLightbulb />,
    label: 'Tip',
    border: 'border-emerald-500/60 dark:border-emerald-400/50',
    bg: 'bg-emerald-50 dark:bg-emerald-950/30',
    headerText: 'text-emerald-800 dark:text-emerald-300',
  },
  warning: {
    icon: <IconAlertTriangle />,
    label: 'Warning',
    border: 'border-amber-500/60 dark:border-amber-400/50',
    bg: 'bg-amber-50 dark:bg-amber-950/30',
    headerText: 'text-amber-800 dark:text-amber-300',
  },
  caution: {
    icon: <IconStop />,
    label: 'Caution',
    border: 'border-red-500/60 dark:border-red-400/50',
    bg: 'bg-red-50 dark:bg-red-950/30',
    headerText: 'text-red-800 dark:text-red-300',
  },
  important: {
    icon: <IconBroadcast />,
    label: 'Important',
    border: 'border-purple-500/60 dark:border-purple-400/50',
    bg: 'bg-purple-50 dark:bg-purple-950/30',
    headerText: 'text-purple-800 dark:text-purple-300',
  },
  warm: {
    icon: <IconFlame />,
    label: 'Warm',
    border: 'border-orange-500/60 dark:border-orange-400/50',
    bg: 'bg-orange-50 dark:bg-orange-950/30',
    headerText: 'text-orange-800 dark:text-orange-300',
  },
};

/* ── Component ─────────────────────────────────────────── */

export default function Callout({ type, children }: CalloutProps) {
  const { icon, label, border, bg, headerText } = CONFIG[type];

  return (
    <div
      className={`callout my-4 rounded-md border-l-[3px] p-4 ${border} ${bg}`}
    >
      <div className={`callout-header mb-1.5 flex items-center gap-1.5 text-[13px] font-semibold ${headerText}`}>
        <span className="flex-shrink-0">{icon}</span>
        <span>{label}</span>
      </div>
      <div className="callout-body text-sm">{children}</div>
    </div>
  );
}
