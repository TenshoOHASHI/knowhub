# knowhub - Product Specification

## Overview
技術ナレッジベースプラットフォーム。
自身の学習記録をWikiとして蓄積し、そのプロジェクト自体がポートフォリオとなる。
バックエンドはGo（マイクロサービス + gRPC + CQRS）、フロントエンドはNext.js。

## Concept
- **技術書を書く感覚**で学習内容をアウトプットする場所
- このプロジェクト自体が最大のポートフォリオ作品
- AI活用による知識検索・要約などの面白い機能を追加予定

## Target User
- 自分自身（学習記録の管理・検索）
- 採用担当者（技術力・継続力の可視化）
- AIアシスタント（MCP経由でナレッジにアクセス）

## Tech Stack

### Backend
- Language: Go 1.22+
- Architecture: Microservices
- Communication: gRPC (inter-service), REST (external)
- Auth: JWT (RS256)
- Patterns: CQRS (Wiki Service)
- AI: OpenAI API / Local LLM (Phase 9)
- MCP: Model Context Protocol Server (Phase 10)

### Frontend
- Framework: Next.js 14+ (App Router)
- Language: TypeScript
- Styling: Tailwind CSS

### Infrastructure
- Container: Docker + Docker Compose
- DB: MySQL 8 (Write), Redis 7 (Read Cache)
- Proto: Protocol Buffers 3
- Deploy: VPS (Hostinger等)
- CI/CD: GitHub Actions

## Services

| Service | Port | Role |
|---------|------|------|
| API Gateway | :8080 (REST) | JWT validation, routing, CORS |
| Auth Service | :50051 (gRPC) | Register, Login, JWT |
| Wiki Service | :50052 (gRPC) | Article CRUD (CQRS) |
| Profile Service | :50053 (gRPC) | Self-intro, portfolio items |
| AI Service | :50054 (gRPC) | Article search, summarization (Phase 9) |
| MCP Server | :5005 (stdio/SSE) | AI assistant integration (Phase 10) |

## Database Design

### MySQL (Auth Service)
- users table

### MySQL (Wiki Service)
- articles table

### MySQL (Profile Service)
- profiles table
- portfolio_items table

### Redis (Wiki Service - Read Model)
- article:{id} → cached article JSON
- articles:list → cached article list

## Development Phases

### Phase 1: Project Setup ✅
- [x] Create project directory
- [x] Docker Compose setup (MySQL, Redis)
- [x] Go module init for each service
- [x] Proto definitions
- [x] Generate Go code from proto

### Phase 2: Wiki Service (Basic CRUD) ✅
- [x] DB connection & migration
- [x] Repository layer
- [x] gRPC handler (Create, Read, Update, Delete)
- [x] Unit tests (model, handler with mock)

### Phase 3: Auth Service + JWT ✅
- [x] User registration (bcrypt)
- [x] Login & JWT generation (RS256)
- [x] RSA key pair generation
- [x] gRPC Reflection

### Phase 4: API Gateway ✅
- [x] REST → gRPC routing (Go 1.22 ServeMux)
- [x] Wiki REST endpoints
- [x] Auth REST endpoints

### Phase 5: CQRS Implementation (Wiki) ✅
- [x] Command/Query repository separation
- [x] Redis read model with fallback to MySQL
- [x] Cache TTL (10 minutes)

### Phase 6: Profile Service (In Progress)
- [ ] Proto definition & code generation
- [ ] Profile model & repository
- [ ] Portfolio item model & repository
- [ ] gRPC handler
- [ ] main.go with service registration

### Phase 7: Frontend (Next.js)
- [ ] Project setup
- [ ] Wiki page (article list + detail)
- [ ] Admin page (article editor)
- [ ] Profile page
- [ ] Portfolio page (developing / completed state)
- [ ] Dark mode

### Phase 8: Polish + Deploy (VPS)
- [ ] VPS setup (SSH, firewall)
- [ ] Nginx + Let's Encrypt
- [ ] Docker Compose production config
- [ ] GitHub Actions CI/CD
- [ ] slog structured logging
- [ ] Error response cleanup (hide internals)

### Phase 9: AI Service
- [ ] Article search with vector embeddings
- [ ] Article summarization
- [ ] Q&A based on wiki content
- [ ] Chat interface

### Phase 10: MCP Server
- [ ] MCP Server implementation (Go)
- [ ] Tools: create_article, search_articles, list_articles
- [ ] Resources: article content access
- [ ] Integration with Claude Desktop / other AI assistants
