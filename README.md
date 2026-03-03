# EAA Recruit (FastAPI + React, PostgreSQL)

SRS/SDS-aligned recruitment platform foundation with:

- PostgreSQL-backed user/job/application data
- Role-based access: `candidate`, `recruiter`, `administrator`, `super_admin`
- JWT auth + bcrypt + lockout policy + rate limiting
- Candidate application flow with CV upload
- Recruiter job management + ranking view
- Admin user management endpoints
- Local TF-IDF/Cosine-style scoring adapter with future FastAPI migration path

## Stack

- Backend: Python, FastAPI, SQLAlchemy, PostgreSQL
- Frontend: React + Vite
- Deployment: Docker Compose

## Quick Start (Docker)

```bash
docker-compose up -d --build
```

Services:

- API: `http://localhost:8080`
- Health: `http://localhost:8080/healthz`
- Postgres: `localhost:5432`

## Local Start

1. Backend:

```bash
pip install -r requirements.txt
uvicorn backend_fastapi.main:app --reload --host 0.0.0.0 --port 8080
```

2. Frontend:

```bash
cd frontend
npm install
npm run dev
```

Frontend URL: `http://localhost:3000`

## Environment Variables

- `JWT_SECRET`
- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `DB_SSLMODE`
- `CORS_ORIGIN` (default `http://localhost:3000`)

## API Base Paths

Both are available:

- `/api/v1/*` (preferred)
- `/*` (backward-compatible mirror)

## Key Endpoints

Public:

- `POST /api/v1/signup`
- `POST /api/v1/login`
- `GET /api/v1/jobs`
- `GET /api/v1/jobs/:id`

Protected candidate:

- `GET /api/v1/users`
- `DELETE /api/v1/users/me`
- `GET /api/v1/applications`
- `GET /api/v1/applications/:id`
- `POST /api/v1/applications` (multipart with `cv`)

Recruiter/Admin:

- `POST /api/v1/jobs`
- `PUT /api/v1/jobs/:id`
- `POST /api/v1/jobs/:id/publish`
- `POST /api/v1/jobs/:id/close`
- `POST /api/v1/jobs/:id/archive`
- `GET /api/v1/jobs/:jobId/ranking`
- `GET /api/v1/applications/:id/explainability`

Admin:

- `GET /api/v1/admin/users`
- `PATCH /api/v1/admin/users/:id/role`
- `PATCH /api/v1/admin/users/:id/status`

## Default Super Admin

- Email: `admin@admin.admin`
- Password: `CqZP99nfbUI2M#3`

Created/updated automatically at startup.

## Backup Scripts

- PowerShell: `scripts/backup-postgres.ps1`
- Shell: `scripts/backup-postgres.sh`
