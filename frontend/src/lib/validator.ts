// ============================================================
// TypeScript バリデーションライブラリ — ゼロから学ぶ実装
// ============================================================
// 学べる概念: クロージャ / メソッドチェーン / ジェネリクス
//             ビルダーパターン / Result型 / 判別可能なユニオン
// ============================================================

// ----- 5. Result型（判別可能なユニオン）-----
// success フィールドで型を絞り込める（Discriminated Union）
export type Result<T> =
  | { success: true; data: T; errors: [] }
  | { success: false; errors: string[] };

// ----- ルールの型 -----
// Rule<T> = 「T型の値を受け取って、成功ならnull、失敗ならエラーメッセージを返す」関数
type Rule<T> = (value: T) => string | null;

// ============================================================
// Validator<T> — ジェネリックベースクラス
// ============================================================
// T = バリデーション対象の型（string, number など）
export class Validator<T> {
  // 1. クロージャ: この配列にルールが蓄積される
  //    外側の関数（インスタンス）が生きている限り、この配列も生きる
  protected rules: Rule<T>[] = [];

  // 2. メソッドチェーン: ルールを追加して this（自分自身）を返す
  //    返り値が同じ Validator<T> なので、.addRule().addRule() と繋げる
  protected addRule(rule: Rule<T>): this {
    this.rules.push(rule);
    return this; // ← チェーンの鍵: 自分自身を返す
  }

  // 4. ビルダーパターン: 「構築（addRule）」と「実行（validate）」を分ける
  //    蓄積した rules をすべて実行して Result 型を返す
  validate(value: T): Result<T> {
    const errors: string[] = [];

    for (const rule of this.rules) {
      // 各ルールを実行 → null なら合格、文字列ならエラーメッセージ
      const error = rule(value);
      if (error !== null) {
        errors.push(error);
      }
    }

    // 5. Result型: 例外を使わず、結果を値で表現
    if (errors.length === 0) {
      return { success: true, data: value, errors: [] };
    }
    return { success: false, errors };
  }
}

// ============================================================
// StringValidator — Validator<string> を拡張
// ============================================================
export class StringValidator extends Validator<string> {
  // 必須チェック（空文字を弾く）
  required(message = '必須項目です'): this {
    return this.addRule((value) =>
      value.trim().length === 0 ? message : null,
    );
  }

  // 最小文字数
  min(length: number, message?: string): this {
    return this.addRule((value) =>
      value.length < length
        ? message || `${length}文字以上で入力してください`
        : null,
    );
  }

  // 最大文字数
  max(length: number, message?: string): this {
    return this.addRule((value) =>
      value.length > length
        ? message || `${length}文字以内で入力してください`
        : null,
    );
  }

  // 正規表現マッチ
  regex(pattern: RegExp, message: string): this {
    return this.addRule((value) => (!pattern.test(value) ? message : null));
  }

  // メールアドレス形式
  email(message = 'メールアドレスの形式が正しくありません'): this {
    return this.addRule((value) =>
      !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value) ? message : null,
    );
  }

  // カスタムルール（任意のチェック関数）
  custom(check: (value: string) => boolean, message: string): this {
    return this.addRule((value) => (check(value) ? null : message));
  }
}

// ============================================================
// NumberValidator — Validator<number> を拡張
// ============================================================
export class NumberValidator extends Validator<number> {
  required(message = '数値を入力してください'): this {
    return this.addRule((value) => (isNaN(value) ? message : null));
  }

  min(n: number, message?: string): this {
    return this.addRule((value) =>
      value < n ? message || `${n}以上の数値を入力してください` : null,
    );
  }

  max(n: number, message?: string): this {
    return this.addRule((value) =>
      value > n ? message || `${n}以下の数値を入力してください` : null,
    );
  }

  integer(message = '整数で入力してください'): this {
    return this.addRule((value) => (!Number.isInteger(value) ? message : null));
  }
}

// ============================================================
// ファクトリ関数 — エントリーポイント
// ============================================================
// 3. ジェネリクス: 戻り値の型が StringValidator と決まるので、
//    TypeScriptは以降のメソッドチェーンで string 型を追跡できる
export function string(): StringValidator {
  return new StringValidator();
}

export function number(): NumberValidator {
  return new NumberValidator();
}

// ============================================================
// 使い方
// ============================================================
// 注意:
//   ライブラリファイルの直下でサンプルコードを実行すると、
//   Next.js の production build 中にも実行されてログが出ます。
//   そのため、ここでは「実行されない例」としてコメントに残します。
//
// パスワードバリデーション例:
//   const result = string()
//     .required('パスワードを入力してください')
//     .min(8, 'パスワードは8文字以上必要です')
//     .max(128)
//     .regex(/[A-Z]/, '大文字を1つ以上含めてください')
//     .regex(/[a-z]/, '小文字を1つ以上含めてください')
//     .regex(/[0-9]/, '数字を1つ以上含めてください')
//     .validate('Password1');
//
// Result型の使い方:
//   if (result.success) {
//     // result.data は string として扱える
//   } else {
//     // result.errors は string[] として扱える
//   }
