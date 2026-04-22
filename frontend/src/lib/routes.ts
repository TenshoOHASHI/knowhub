import { FiBook, FiUser, FiBriefcase, FiEdit } from 'react-icons/fi';
import { IconType } from 'react-icons';
import { VscVersions } from 'react-icons/vsc';

// ルートの中身を定義
export interface Route {
  label: string;
  href: string;
  icon: IconType;
}

// 複数のルートリンクを管理
export const NAV_LINKS: Route[] = [
  { label: 'Wiki', href: '/wiki', icon: FiBook },
  { label: 'Admin', href: '/admin', icon: FiEdit },
  { label: 'Profile', href: '/profile', icon: FiUser },
  { label: 'Portfolio', href: '/portfolio', icon: FiBriefcase },
  { label: 'Log', href: '/changelog', icon: VscVersions },
];
