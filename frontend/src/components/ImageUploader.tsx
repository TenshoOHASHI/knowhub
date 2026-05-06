'use client';

import { uploadImage } from '@/lib/api';
import { useRef, useState } from 'react';

interface ImageUploaderProps {
  value: string;
  onChange: (url: string) => void;
  label?: string;
  shape?: 'circle' | 'rectangle';
}

const MAX_SIZE = 5 * 1024 * 1024; // 5MB
const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];

export default function ImageUploader({
  value,
  onChange,
  label = '画像',
  shape = 'rectangle',
}: ImageUploaderProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleFile = async (file: File) => {
    setError(null);

    if (!ALLOWED_TYPES.includes(file.type)) {
      setError('JPEG, PNG, GIF, WebP のみアップロードできます');
      return;
    }
    if (file.size > MAX_SIZE) {
      setError('ファイルサイズは5MB以下にしてください');
      return;
    }

    setUploading(true);
    try {
      const url = await uploadImage(file);
      onChange(url);
    } catch {
      setError('アップロードに失敗しました');
    } finally {
      setUploading(false);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    const file = e.dataTransfer.files[0];
    if (file) handleFile(file);
  };

  const isCircle = shape === 'circle';

  return (
    <div>
      <label className='block mb-1'>{label}</label>
      <div
        onClick={() => inputRef.current?.click()}
        onDragOver={(e) => {
          e.preventDefault();
          setDragOver(true);
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={handleDrop}
        className={`
          relative cursor-pointer border-2 border-dashed transition-colors
          ${dragOver ? 'border-green-500 bg-green-500/10' : 'border-stone-400 dark:border-stone-600 hover:border-stone-500'}
          flex items-center justify-center
          ${isCircle ? 'w-32 h-32 rounded-full overflow-hidden' : 'w-full h-40 rounded-lg overflow-hidden'}
        `}
      >
        {value ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={value}
            alt='preview'
            className={`${isCircle ? 'w-full h-full object-cover' : 'max-h-full object-contain'}`}
          />
        ) : (
          <span className='text-sm text-stone-500'>
            {uploading
              ? 'アップロード中...'
              : 'クリックまたはドラッグ&ドロップ'}
          </span>
        )}
      </div>
      <input
        ref={inputRef}
        type='file'
        accept='image/jpeg,image/png,image/gif,image/webp'
        className='hidden'
        onChange={(e) => {
          const file = e.target.files?.[0];
          if (file) handleFile(file);
          // Reset so the same file can be re-selected
          e.target.value = '';
        }}
      />
      {error && <p className='text-red-500 text-sm mt-1'>{error}</p>}
    </div>
  );
}
