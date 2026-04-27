'use client';

import { FiX } from 'react-icons/fi';

interface ConfirmModalProps {
  isOpen: boolean;
  title: string;
  message: string;
  onConfirm: () => void;
  onCancel: () => void;
}

export default function ConfirmModal({
  isOpen,
  title,
  message,
  onConfirm,
  onCancel,
}: ConfirmModalProps) {
  if (!isOpen) return null;

  return (
    // 背景（暗いオーバーレイ）、画面全体に被せる（一番手前に表示）
    <div
      className='fixed inset-0 z-50 bg-black/50 flex items-center justify-center'
      onClick={onCancel} // 背景クリックで閉じる
    >
      {/* モーダル本体 */}
      <div
        className='bg-white dark:bg-stone-800 rounded-lg p-6 max-w-md w-full mx-4 shadow-xl'
        onClick={(e) => e.stopPropagation()} // モーダル内クリックは背景に伝播させない
      >
        <div className='flex justify-between items-center mb-4'>
          <h3 className='text-lg font-semibold'>{title}</h3>
          <button onClick={onCancel}>
            <FiX />
          </button>
        </div>
        <p className='text-gray-600 dark:text-stone-400 mb-6'>{message}</p>
        <div className='flex justify-end gap-3'>
          <button
            type="button"
            onClick={onCancel}
            className='px-4 py-2 border border-black dark:border-stone-600 rounded-lg'
          >
            キャンセル
          </button>
          <button
            type="button"
            onClick={onConfirm}
            className='px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700'
          >
            削除
          </button>
        </div>
      </div>
    </div>
  );
}
