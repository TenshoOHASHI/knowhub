'use client';
import { useAuth } from '@/context/AuthContext';
import { useToast } from '@/context/ToastContext';
import { useRouter } from 'next/navigation';
import { useState } from 'react';

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const { toast } = useToast();
  const { login } = useAuth();

  async function handleSubmit(e: React.SubmitEvent<HTMLFormElement>) {
    e.preventDefault();
    setError('');

    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });

      if (!res.ok) {
        setError('ログインに失敗しました');
        return;
      }

      toast('ログインに成功しました。', 'success');
      login(); // AuthContext の isLoggedIn を即座に true にする
      router.push('/admin');
    } catch {
      setError('エラーが発生しました');
    }
  }

  return (
    <div className='min-h-2/3 flex items-center justify-center'>
      <form
        onSubmit={handleSubmit}
        className='w-full max-w-sm space-y-4 p-6 bg-white dark:bg-stone-800 rounded-lg shadow'
      >
        <h1>Login</h1>
        {error && <p className='text-xl font-bold text-center'>{error}</p>}
        <input
          type='email'
          placeholder='Email'
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          className='w-full px-3 py-2 border rounded dark:bg-stone-700 dark:border-stone-600'
          autoComplete='username'
        />
        <input
          type='password'
          placeholder='******'
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          className='w-full px-3 py-2 border rounded dark:bg-stone-700 dark:border-stone-600'
          autoComplete='current-password'
        />
        <button
          type='submit'
          className='w-full p-2 border rounded dark:bg-stone-700 dark:border-stone-600'
        >
          ログイン
        </button>
      </form>
    </div>
  );
}
