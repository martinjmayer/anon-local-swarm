# Stack Evaluation Summary

> Compact reference for stack selection decisions. For full evaluations, see `standards/reference/`.

## Backend Tier Rankings

| Rank | Stack | Agent Suitability | First-Pass Rate | Typical Iterations |
|------|-------|-------------------|----------------:|-------------------:|
| 1 | Go + Gin/Chi/Echo | ★★★★ | 70-80% | 1-3 |
| 2 | Node.js + Fastify/Hono (TS) | ★★★★ | 65-75% | 2-4 |
| 3 | Java + Spring Boot | ★★★★ | 65-75% | 2-4 |
| 4 | C# + ASP.NET Core | ★★★★ | 65-75% | 2-4 |
| 5 | Python + FastAPI | ★★★ | 60-70% | 3-5 |
| 6 | Rust + Axum | ★★★ | 40-55% | 5-10 |

## Frontend Tier Rankings

| Rank | Framework | Agent Suitability | First-Pass Rate | Typical Iterations |
|------|-----------|-------------------|----------------:|-------------------:|
| 1 | React 19 + Vite + TS | ★★★★★ | 70-80% | 2-4 |
| 2 | Next.js 15+ | ★★★★ | 60-70% | 3-5 |
| 3 | Vue 3 + Nuxt 3 | ★★★★ | 60-70% | 3-5 |
| 4 | Angular 18+ | ★★★ | 55-65% | 4-6 |
| 5 | SvelteKit (Svelte 5) | ★★★ | 50-60% | 4-6 |

## Key Decision Guide

- **Maximum SDK coverage:** Node.js + TypeScript
- **Maximum correctness guarantees:** Go (pragmatic) or Rust (strict)
- **Enterprise ecosystem:** Java + Spring Boot
- **Frontend default:** React 19 + Vite + TypeScript + shadcn/ui
- **Full-stack SSR:** Next.js 15+

> **Full evaluations:** [Backend](reference/backend-stack-evaluation.md) | [Frontend](reference/frontend-stack-evaluation.md)
