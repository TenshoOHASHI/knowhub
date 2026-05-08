'use client';

import { useToast } from '@/context/ToastContext';
import { useEffect, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import Link from 'next/link';
import type { Category } from '@/lib/types';
import {
  getCategories,
  getArticle,
  saveArticle,
  deleteArticle,
  uploadImage,
} from '@/lib/api';
import EditorPreview from './EditorPreview';
import { useRouter } from 'next/navigation';
import ConfirmModal from './ConfirmModal';
import { FiEdit3 } from 'react-icons/fi';

interface EditorProps {
  embedded?: boolean;
}

export default function Editor({ embedded = false }: EditorProps) {
  // URLパラメータ ?id=xxx があれば更新モード
  const searchParams = useSearchParams();
  const editId = searchParams.get('id');

  const [message, setMessage] = useState<string | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const { toast } = useToast();
  const [categories, setCategories] = useState<Category[]>([]);

  // フォームの初期値（更新モード時は既存データで埋める）
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [categoryId, setCategoryId] = useState('');
  const [visibility, setVisibility] = useState('public');
  const [saving, setSaving] = useState(false);

  const router = useRouter();

  // 更新モード: 既存記事データを取得してフォームにセット
  useEffect(() => {
    if (!editId) return;
    getArticle(editId)
      .then((data) => {
        const article = data.Article;
        setTitle(article.title);
        setContent(article.content);
        setCategoryId(article.category_id || '');
        setVisibility(article.visibility || 'public');
      })
      .catch(() => toast('記事の取得に失敗しました', 'error'));
  }, [editId, toast]);

  // カテゴリ一覧を取得
  useEffect(() => {
    getCategories()
      .then((data) => setCategories(data.categories || []))
      .catch(() => {});
  }, []);

  // 子コンポーネント（EditorPreview）から呼ばれるコールバック
  const handleContentChange = (value: string) => {
    setContent(value);
  };

  // 画像アップロード
  const handleImageUpload = () => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = 'image/jpeg,image/png,image/gif,image/webp';
    input.onchange = async () => {
      const file = input.files?.[0];
      if (!file) return;
      try {
        const url = await uploadImage(file);
        setContent((prev) => `${prev}\n![${file.name}](${url})\n`);
      } catch {
        toast('画像のアップロードに失敗しました', 'error');
      }
    };
    input.click();
  };

  // 作成・更新
  const handleSubmit = async (e: React.SubmitEvent) => {
    e.preventDefault();
    if (saving) return;
    setSaving(true);
    try {
      await saveArticle({
        id: editId || undefined,
        title,
        content,
        category_id: categoryId,
        visibility,
      });
      const msg = editId ? '記事を更新しました' : '記事を作成しました';
      toast(msg, 'success');
      setMessage(msg);
    } catch {
      const msg = editId ? '更新に失敗しました' : '作成に失敗しました';
      toast(msg, 'error');
      setMessage(msg);
    } finally {
      setSaving(false);
    }
  };

  // 削除処理
  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await deleteArticle(deleteTarget);
      toast('記事を削除しました', 'success');
      setDeleteTarget(null);
      router.push('/wiki');
    } catch {
      toast('削除に失敗しました', 'error');
    }
  };

  return (
    <>
      {/* タイトルヘッダー - embeddedモードでは非表示 */}
      {!embedded && (
        <div className='max-w-7xl mx-auto p-6'>
          <div className='mb-6'>
            <div className='flex items-center gap-3'>
              <div className='w-10 h-10 rounded-xl bg-gradient-to-br from-amber-500 to-orange-600 flex items-center justify-center shadow-lg'>
                <FiEdit3 size={20} className='text-white' />
              </div>
              <div>
                <h1 className='text-2xl font-bold text-stone-900 dark:text-stone-100'>
                  {editId ? '記事編集' : '記事エディタ'}
                </h1>
                <p className='text-sm text-stone-500 dark:text-stone-400'>
                  Wiki の記事を作成・編集します
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      <form
        onSubmit={handleSubmit}
        className={embedded ? 'space-y-6 p-6' : 'space-y-6'}
      >
        {/* 更新モード時のみhidden inputでIDを送信 */}
        {editId && <input type='hidden' name='id' value={editId} />}

        {/* 入力エリア */}
        <div className='rounded-xl border border-stone-200 dark:border-stone-700 bg-white dark:bg-stone-800 shadow-sm p-6 space-y-5'>
          <div>
            <label className='block text-sm font-semibold mb-2 text-stone-900 dark:text-stone-100'>
              タイトル
            </label>
            <input
              name='title'
              type='text'
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className='w-full border border-stone-300 dark:border-stone-600 rounded-xl px-4 py-2.5 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition'
              placeholder='記事タイトルを入力してください'
            />
          </div>

          <div className='grid grid-cols-1 md:grid-cols-2 gap-5'>
            <div>
              <label className='block text-sm font-semibold mb-2 text-stone-900 dark:text-stone-100'>
                カテゴリ
              </label>
              <select
                name='category_id'
                value={categoryId}
                onChange={(e) => setCategoryId(e.target.value)}
                className='w-full border border-stone-300 dark:border-stone-600 rounded-xl px-4 py-2.5 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition'
              >
                <option value=''>カテゴリなし</option>
                {categories.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className='block text-sm font-semibold mb-2 text-stone-900 dark:text-stone-100'>
                公開設定
              </label>
              <select
                name='visibility'
                value={visibility}
                onChange={(e) => setVisibility(e.target.value)}
                className='w-full border border-stone-300 dark:border-stone-600 rounded-xl px-4 py-2.5 bg-white dark:bg-stone-900 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition'
              >
                <option value='public'>一般公開</option>
                <option value='locked'>限定公開</option>
              </select>
            </div>
          </div>
        </div>

        {/* 子コンポーネントに content と onContentChange を渡す */}
        <EditorPreview
          content={content}
          onContentChange={handleContentChange}
        />

        {/* 画像アップロードボタン */}
        <button
          type='button'
          onClick={handleImageUpload}
          className='inline-flex items-center gap-2 px-4 py-2.5 rounded-xl border border-stone-300 dark:border-stone-600 hover:bg-stone-50 dark:hover:bg-stone-800 text-sm font-medium text-stone-700 dark:text-stone-300 transition-colors'
        >
          <svg
            className='w-4 h-4'
            fill='none'
            stroke='currentColor'
            viewBox='0 0 24 24'
          >
            <path
              strokeLinecap='round'
              strokeLinejoin='round'
              strokeWidth={2}
              d='M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z'
            />
          </svg>
          画像をアップロード
        </button>

        {/* ボタンエリア */}
        <div className='flex flex-wrap gap-3'>
          <button
            type='submit'
            disabled={saving}
            className='bg-stone-900 text-white dark:bg-stone-100 dark:text-stone-900 px-6 py-2.5 rounded-lg text-sm font-medium hover:bg-stone-800 dark:hover:bg-stone-200 transition-colors disabled:opacity-50 disabled:cursor-not-allowed'
          >
            {saving ? '保存中...' : editId ? '更新' : '作成'}
          </button>

          {editId && (
            <button
              type='button'
              onClick={() => setDeleteTarget(editId)}
              className='px-6 py-2.5 rounded-lg border border-stone-300 dark:border-stone-600 text-stone-700 dark:text-stone-300 text-sm font-medium hover:bg-stone-100 dark:hover:bg-stone-800 transition-colors'
            >
              削除
            </button>
          )}
        </div>

        {/* 削除idが存在する場合、モーダルを開く*/}
        {deleteTarget && (
          <ConfirmModal
            isOpen={true}
            title='記事削除'
            message={`この記事を削除しますか？`}
            onConfirm={() => handleDelete()}
            onCancel={() => setDeleteTarget(null)}
          />
        )}
      </form>

      {/* 作成・更新成功時にWikiへのリンクを表示 */}
      {message &&
        (message.includes('記事を更新しました') ||
          message?.includes('記事を作成しました')) && (
          <div
            className={`p-4 bg-stone-50 dark:bg-stone-800/50 border border-stone-200 dark:border-stone-700 rounded-lg flex items-center gap-3 shadow-sm ${!embedded ? 'mt-6' : 'mt-4 mx-6'}`}
          >
            <svg
              className='w-5 h-5 text-stone-600 dark:text-stone-400'
              fill='none'
              stroke='currentColor'
              viewBox='0 0 24 24'
            >
              <path
                strokeLinecap='round'
                strokeLinejoin='round'
                strokeWidth={2}
                d='M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z'
              />
            </svg>
            <span className='text-stone-800 dark:text-stone-200 font-medium'>
              {message}
            </span>
            <Link
              href='/wiki'
              className='ml-auto text-sm text-stone-600 dark:text-stone-400 underline hover:text-stone-900 dark:hover:text-stone-200 transition-colors'
            >
              Wikiで確認 →
            </Link>
          </div>
        )}
    </>
  );
}
