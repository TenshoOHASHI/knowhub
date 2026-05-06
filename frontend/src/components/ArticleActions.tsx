'use client';

import { useState, useEffect, useCallback } from 'react';
import { FiHeart, FiBookmark } from 'react-icons/fi';
import { getFingerprint } from '@/lib/fingerprint';
import { toggleLike, getLikeCount, saveArticleBookmark, unsaveArticleBookmark, isArticleSaved } from '@/lib/api';

interface ArticleActionsProps {
  articleId: string;
}

export default function ArticleActions({ articleId }: ArticleActionsProps) {
  const [likeCount, setLikeCount] = useState(0);
  const [liked, setLiked] = useState(false);
  const [saved, setSaved] = useState(false);
  const [loading, setLoading] = useState(true);
  const [likePending, setLikePending] = useState(false);
  const [savePending, setSavePending] = useState(false);

  useEffect(() => {
    async function load() {
      try {
        const fp = await getFingerprint();
        const [likeData, savedData] = await Promise.all([
          getLikeCount(articleId, fp),
          isArticleSaved(articleId, fp),
        ]);
        setLikeCount(likeData.count);
        setLiked(likeData.liked);
        setSaved(savedData.saved);
      } catch {
        // Silently fail — non-critical feature
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [articleId]);

  const handleLike = useCallback(async () => {
    if (likePending) return;
    setLikePending(true);
    try {
      const fp = await getFingerprint();
      const data = await toggleLike(articleId, fp);
      setLikeCount(data.count);
      setLiked(data.liked);
    } catch {
      // Silently fail
    } finally {
      setLikePending(false);
    }
  }, [articleId, likePending]);

  const handleSave = useCallback(async () => {
    if (savePending) return;
    setSavePending(true);
    try {
      const fp = await getFingerprint();
      if (saved) {
        await unsaveArticleBookmark(articleId, fp);
        setSaved(false);
      } else {
        await saveArticleBookmark(articleId, fp);
        setSaved(true);
      }
    } catch {
      // Silently fail
    } finally {
      setSavePending(false);
    }
  }, [articleId, saved, savePending]);

  if (loading) {
    return (
      <div className="flex items-center gap-3">
        <div className="w-16 h-8 bg-stone-100 dark:bg-stone-800 rounded-full animate-pulse" />
        <div className="w-8 h-8 bg-stone-100 dark:bg-stone-800 rounded-full animate-pulse" />
      </div>
    );
  }

  return (
    <div className="flex items-center gap-3">
      {/* Like button */}
      <button
        onClick={handleLike}
        disabled={likePending}
        className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm font-medium transition-all duration-200 ${
          liked
            ? 'bg-rose-100 text-rose-600 dark:bg-rose-900/30 dark:text-rose-400'
            : 'bg-stone-100 text-stone-500 hover:bg-stone-200 dark:bg-stone-800 dark:text-stone-400 dark:hover:bg-stone-700'
        } disabled:opacity-60 disabled:cursor-not-allowed`}
        title={liked ? 'いいね解除' : '参考になった！'}
      >
        <FiHeart
          size={14}
          className={liked ? 'fill-current' : ''}
        />
        <span>{likeCount}</span>
      </button>

      {/* Save button */}
      <button
        onClick={handleSave}
        disabled={savePending}
        className={`inline-flex items-center justify-center w-8 h-8 rounded-full transition-all duration-200 ${
          saved
            ? 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400'
            : 'bg-stone-100 text-stone-500 hover:bg-stone-200 dark:bg-stone-800 dark:text-stone-400 dark:hover:bg-stone-700'
        } disabled:opacity-60 disabled:cursor-not-allowed`}
        title={saved ? '保存解除' : '保存する'}
      >
        <FiBookmark
          size={14}
          className={saved ? 'fill-current' : ''}
        />
      </button>
    </div>
  );
}
