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

export default function Editor() {
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
    <div className='max-w-7xl mx-auto p-6'>
      <h1 className='text-3xl font-bold mb-4'>
        {editId ? '記事編集' : '記事エディタ'}
      </h1>
      <form onSubmit={handleSubmit} className='space-y-4'>
        {/* 更新モード時のみhidden inputでIDを送信 */}
        {editId && <input type='hidden' name='id' value={editId} />}

        {/* 入力エリア */}
        <div>
          <label className='block text-sm font-medium mb-1'>タイトル</label>
          <input
            name='title'
            type='text'
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 bg-transparent mb-2'
            placeholder='記事タイトル'
          />

          <label className='block text-sm font-medium mb-2'>カテゴリ</label>
          <select
            name='category_id'
            value={categoryId}
            onChange={(e) => setCategoryId(e.target.value)}
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 bg-transparent'
          >
            <option value=''>カテゴリなし</option>
            {categories.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>

          <label className='block text-sm font-medium mb-2 mt-2'>
            公開設定
          </label>
          <select
            name='visibility'
            value={visibility}
            onChange={(e) => setVisibility(e.target.value)}
            className='w-full border border-black dark:border-stone-600 rounded-lg p-2 bg-transparent'
          >
            <option value='public'>一般公開</option>
            <option value='locked'>限定公開</option>
          </select>
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
          className='px-3 py-1.5 rounded-lg border border-stone-400 dark:border-stone-600 hover:bg-stone-100 dark:hover:bg-stone-800 text-sm block'
        >
          画像 +
        </button>

        {/* ボタン: IDがあれば「更新」（緑）、なければ「作成」（黒） */}
        <button
          type='submit'
          disabled={saving}
          className={`px-4 py-2 rounded-lg ${
            editId
              ? 'bg-green-700 dark:bg-green-300 text-white dark:text-stone-900 hover:bg-green-800 dark:hover:bg-green-300/90'
              : 'bg-stone-900 dark:bg-stone-100 text-white dark:text-stone-900 hover:bg-stone-800 dark:hover:bg-stone-200'
          }`}
        >
          {saving ? '保存中...' : editId ? '更新' : '作成'}
        </button>

        {editId && (
          <button
            type='button'
            onClick={() => setDeleteTarget(editId)}
            className='px-4 py-2 rounded-lg bg-red-400 text-whited hover:dark:bg-red-400/90 hover:bg-red-400/90 ml-3 hover:text-stone-300'
          >
            削除
          </button>
        )}

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
          <div className='mt-3 flex items-center gap-2 text-sm text-green-600 dark:text-green-400'>
            <span>{message}</span>
            <Link
              href='/wiki'
              className='underline hover:text-green-800 dark:hover:text-green-300 '
            >
              Wikiで確認 →
            </Link>
          </div>
        )}
    </div>
  );
}
