# My Portfolio - Product Specification

## Overview
個人のポートフォリオ + Wikiサイト。
バックエンドはGo（マイクロサービス + gRPC + CQRS）、フロントエンドはNext.js。

## Target User
- 自分自身（学習記録の管理）
- 採用担当者（スキルの可視化）

## Tech Stack

### Backend
- Language: Go 1.22+
- Architecture: Microservices
- Communication: gRPC (inter-service), REST (external)
- Auth: JWT (RS256)
- Patterns: CQRS (Wiki Service)

### Frontend
- Framework: Next.js 14+ (App Router)
- Language: TypeScript
- Styling: Tailwind CSS

### Infrastructure
- Container: Docker + Docker Compose
- DB: MySQL 8 (Write), Redis 7 (Read Cache)
- Proto: Protocol Buffers 3
- Deploy: VPS (Hostinger等) + Vercel (Frontend)

## Services

| Service | Port | Role |
|---------|------|------|
| API Gateway | :8080 (REST) | JWT validation, routing |
| Auth Service | :50051 (gRPC) | Register, Login, JWT |
| Wiki Service | :50052 (gRPC) | Article CRUD (CQRS) |
| Profile Service | :50053 (gRPC) | Self-intro, portfolio |

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

### Phase 1: Project Setup
- [x] Create project directory
- [ ] Docker Compose setup (MySQL, Redis)
- [ ] Go module init for each service
- [ ] Proto definitions
- [ ] Generate Go code from proto

### Phase 2: Wiki Service (Basic CRUD first)
- [ ] DB connection & migration
- [ ] Repository layer
- [ ] gRPC handler (Create, Read, Update, Delete)
- [ ] Unit tests

### Phase 3: Auth Service + JWT
- [ ] User registration
- [ ] Login & JWT generation
- [ ] JWT middleware

### Phase 4: API Gateway
- [ ] REST → gRPC routing
- [ ] JWT validation middleware
- [ ] CORS

### Phase 5: CQRS Implementation (Wiki)
- [ ] Command/Query separation
- [ ] Event publishing (Go channel)
- [ ] Redis read model sync
- [ ] Read from Redis, fallback to MySQL

### Phase 6: Profile Service
- [ ] Profile CRUD
- [ ] Portfolio items CRUD

### Phase 7: Frontend (Next.js)
- [ ] Project setup
- [ ] Wiki page (article list + detail)
- [ ] Admin page (article editor)
- [ ] Profile page
- [ ] Portfolio page (loading state)

### Phase 8: Deploy
- [ ] VPS setup (SSH, firewall)
- [ ] Nginx + Let's Encrypt
- [ ] Docker Compose production config
- [ ] GitHub Actions CI/CD
