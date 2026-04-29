import { FiBook, FiUser, FiBriefcase, FiEdit } from 'react-icons/fi';
import { IconType } from 'react-icons';
import { VscVersions } from 'react-icons/vsc';
import { FiLogOut } from 'react-icons/fi';
import { MdOutlineSupportAgent } from 'react-icons/md';

// ルートの中身を定義
export interface Route {
  label: string;
  href: string;
  icon: IconType;
}

// 未ログイン用
export const NAV_LINKS: Route[] = [
  { label: 'Wiki', href: '/wiki', icon: FiBook },
  { label: 'Profile', href: '/profile', icon: FiUser },
  { label: 'Portfolio', href: '/portfolio', icon: FiBriefcase },
  { label: 'Chat', href: '/chat', icon: MdOutlineSupportAgent },
  { label: 'Log', href: '/changelog', icon: VscVersions },
];

// ログイン済み用（Admin 追加）
export const NAV_LINKS_WITH_AUTH: Route[] = [
  { label: 'Wiki', href: '/wiki', icon: FiBook },
  { label: 'Admin', href: '/admin', icon: FiEdit },
  { label: 'Profile', href: '/profile', icon: FiUser },
  { label: 'Portfolio', href: '/portfolio', icon: FiBriefcase },
  { label: 'Chat', href: '/chat', icon: MdOutlineSupportAgent },
  { label: 'Log', href: '/changelog', icon: VscVersions },
];

// ログアウトボタン用（特殊扱い）
export const LOGOUT_ROUTE = { label: 'Logout', icon: FiLogOut };
