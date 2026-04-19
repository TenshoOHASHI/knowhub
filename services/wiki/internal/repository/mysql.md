# database/sqlメソッド

メソッド | 用途 | 戻り値
| --- | --- | --- |
| ExecContent| INSERT, UPDATE, DELETE | Result(影響行数等) |
| QueryRowContent | SELECT １件 | Row(１行スキャン)　|
| QueryContext | SELECT 複数件 | Rows(複数行ループ) |


# 注意点
1. sql.ErrNoRows の処理を忘れない（FindByID）
2. rows.Close() を忘れない（FindAll）
3. context.Context を各メソッドに渡す（キャンセル・タイムアウト対応）
4. SQLインジェクション対策: 値は ? プレースホルダーで渡す（文字列連結しない）
