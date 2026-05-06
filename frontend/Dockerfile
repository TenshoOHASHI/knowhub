# ==========================================
# 1. 依存関係インストール用ステージ (deps)
# ==========================================
# node:22-alpine:
#   Node.js 22 が入った軽量Linux imageです。
#   alpineは小さいため、CI/CDや本番imageでよく使われます。
FROM node:22-alpine AS deps

# コンテナ内の作業ディレクトリです。
# この後のCOPY/RUNは基本的に /app を基準に実行されます。
WORKDIR /app

# package.json / package-lock.json だけ先にコピーします。
# 理由:
#   依存関係だけを先にinstallすると、ソースコード変更時でも
#   package-lock.jsonが変わらなければDocker cacheを使いやすいからです。
COPY frontend/package.json frontend/package-lock.json ./

# package-lock.jsonに完全一致する依存関係をインストールします。
# npm install:
#   package.jsonを見て依存を更新することがあり、環境によって差が出やすい。
# npm ci:
#   package-lock.json通りに入れる。CI/Docker build向き。
# 例:
#   ローカルとGitHub Actionsで同じ依存バージョンを使いたいので npm ci を使います。
RUN npm ci

# ==========================================
# 2. Next.js build用ステージ (builder)
# ==========================================
# depsとは別ステージにして、依存インストールとbuildを分けます。
FROM node:22-alpine AS builder

WORKDIR /app

# Next.jsの匿名telemetry送信を無効化します。
# telemetry:
#   Next.jsが利用状況を匿名で収集する仕組みです。
#   アプリの動作には不要なので、CI/build環境では送らない設定にします。
ENV NEXT_TELEMETRY_DISABLED=1

# depsステージで作ったnode_modulesを再利用します。
# ここで再度 npm ci しないため、buildが速くなります。
COPY --from=deps /app/node_modules ./node_modules

# frontendディレクトリ配下のソースコードを /app にコピーします。
# ここに app/, components/, next.config.js, public/ などが入ります。
COPY frontend ./

# Next.jsのproduction buildを実行します。
# next.config.js 側で output: 'standalone' を使うと、
# 実行に必要な最小ファイルが .next/standalone に出力されます。
RUN npm run build

# ==========================================
# 3. 本番実行用ステージ (runner)
# ==========================================
# 最終imageです。
# deps/buildに使った余分なファイルを持ち込まず、実行に必要なものだけ入れます。
FROM node:22-alpine AS runner

WORKDIR /app

# 本番実行モードです。
# コンテナ内部のNode.js/Next.jsが process.env.NODE_ENV として読みます。
# .env.productionの代わりではなく、アプリの実行モードを示す基本環境変数です。
# 例:
#   process.env.NODE_ENV === 'production' なら、本番用最適化やsecure cookie判定に使えます。
ENV NODE_ENV=production

# Next.js telemetryを本番実行時も無効化します。
# build時だけでなく、コンテナ起動中も送信しないようにします。
ENV NEXT_TELEMETRY_DISABLED=1

# Next.js standalone serverがlistenするポートです。
ENV PORT=3000

# 0.0.0.0 にすることで、コンテナ外から接続できるようにします。
# localhostだけだと、コンテナ内からしか見えない場合があります。
ENV HOSTNAME=0.0.0.0

# Linuxユーザー/グループを作ります。
# なぜ必要か:
#   何も指定しないと、コンテナ内のプロセスがrootで動くことがあります。
#   rootで動かすと、アプリに脆弱性があった時にコンテナ内で強い権限を持たれます。
# ここでは nextjs という実行専用ユーザーを作り、あとで USER nextjs に切り替えます。
#
# addgroup:
#   nodejs というグループを作る
# adduser:
#   nextjs というユーザーを作り、nodejsグループに所属させる
RUN addgroup --system --gid 1001 nodejs \
  && adduser --system --uid 1001 nextjs

# public/ はブラウザから /dark.png のように参照される静的ファイルです。
# 現在は Navbar が /dark.png を使っているため、public/ は必要です。
# 未使用だったNext.js初期SVGは削除済みです。
COPY --from=builder /app/public ./public

# .next/standalone:
#   Next.js standalone buildの実行本体です。
#   server.js と必要最小限のnode_modulesが含まれます。
# --chown:
#   コピーしたファイルの所有者を nextjs:nodejs にします。
# なぜ必要か:
#   この後 USER nextjs に切り替えるため、nextjsユーザーが読める所有権にしておきます。
#   root所有のままでも読める場合はありますが、本番では権限を明示する方が安全です。
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./

# .next/static:
#   JS/CSSなどの静的build成果物です。
#   例: ページ表示に必要なJavaScript bundle、CSS、chunkファイル。
#   .next/standaloneには含まれないため、別途コピーします。
#   これを消すと画面は開いてもJS/CSSが404になり、UIが壊れます。
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

# ここから先の実行ユーザーをnextjsに切り替えます。
USER nextjs

# Docker imageのメタ情報として「このアプリは3000番で待ち受ける」と示します。
# 注意:
#   EXPOSEだけでは外部公開されません。
#   外部公開は docker-compose.prod.yml の ports や nginx の proxy設定で決まります。
# 今回:
#   frontend container は 3000番で待ち受け
#   nginx container が frontend:3000 に内部proxy
#   外部公開されるのは nginx の 80/443
EXPOSE 3000

# Next.js standalone buildが生成したserver.jsを起動します。
CMD ["node", "server.js"]
