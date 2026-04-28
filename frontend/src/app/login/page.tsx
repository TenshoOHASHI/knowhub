'use client';
import { useAuth } from '@/context/AuthContext';
import { useToast } from '@/context/ToastContext';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import { string, type Result } from '@/lib/validator';

// バリデータの定義（各フィールドのルール）
const emailValidator = string()
  .required('メールアドレスを入力してください')
  .email();

const passwordValidator = string()
  .required('パスワードを入力してください')
  .min(8, 'パスワードは8文字以上必要です');

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errors, setErrors] = useState<Record<string, string[]>>({});
  const [error, setError] = useState('');
  const { toast } = useToast();
  const { login } = useAuth();

  async function handleSubmit(e: React.SubmitEvent<HTMLFormElement>) {
    e.preventDefault();
    setError('');

    // クライアント側バリデーション
    const emailResult: Result<string> = emailValidator.validate(email);
    const passwordResult: Result<string> = passwordValidator.validate(password);

    const fieldErrors: Record<string, string[]> = {};
    if (!emailResult.success) fieldErrors.email = emailResult.errors;
    if (!passwordResult.success) fieldErrors.password = passwordResult.errors;

    if (Object.keys(fieldErrors).length > 0) {
      setErrors(fieldErrors);
      return;
    }

    setErrors({});

    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });

      if (!res.ok) {
        setError('メールアドレスまたはパスワードが正しくありません');
        return;
      }

      toast('ログインに成功しました。', 'success');
      login();
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
        <h1 className='text-2xl font-bold'>Login</h1>
        {error && <p className='text-red-500 text-sm text-center'>{error}</p>}

        {/* Email */}
        <div>
          <input
            type='email'
            placeholder='Email'
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className={`w-full px-3 py-2 border rounded dark:bg-stone-700 dark:border-stone-600 ${
              errors.email ? 'border-red-500' : 'border-stone-300'
            }`}
            autoComplete='username'
          />
          {errors.email && (
            <p className='text-red-500 text-xs mt-1'>{errors.email[0]}</p>
          )}
        </div>

        {/* Password */}
        <div>
          <input
            type='password'
            placeholder='Password'
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className={`w-full px-3 py-2 border rounded dark:bg-stone-700 dark:border-stone-600 ${
              errors.password ? 'border-red-500' : 'border-stone-300'
            }`}
            autoComplete='current-password'
          />
          {errors.password && (
            <p className='text-red-500 text-xs mt-1'>{errors.password[0]}</p>
          )}
        </div>
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
