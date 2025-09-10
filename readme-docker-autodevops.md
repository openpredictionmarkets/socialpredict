# Docker Auto DevOps

## What I changed
- Backend: made dotenv optional and normalized DB envs; see [`backend/util/getenv.go`](backend/util/getenv.go:1) and [`backend/util/postgres.go`](backend/util/postgres.go:1).
- Backend Dockerfile: multi-stage static build: [`docker/backend/Dockerfile`](docker/backend/Dockerfile:1).
- Frontend: runtime env template + loader: [`frontend/public/env-config.js.template`](frontend/public/env-config.js.template:1) and [`frontend/index.html`](frontend/index.html:29).
- Frontend Dockerfile: build + nginx runtime with envplate and entrypoint: [`docker/frontend/Dockerfile`](docker/frontend/Dockerfile:1), [`docker/frontend/entrypoint.sh`](docker/frontend/entrypoint.sh:1), [`docker/frontend/nginx.conf`](docker/frontend/nginx.conf:1).
- CI: new GitHub Actions workflow that builds and pushes to Docker Hub on branch devops-docker: [`.github/workflows/dockerhub-images.yml`](.github/workflows/dockerhub-images.yml:1).

## Build locally (quick)
1. Build backend:
```bash
docker buildx build --platform linux/amd64 -f docker/backend/Dockerfile -t jmartincufre/socialpredict-backend:local --load .
```
2. Build frontend:
```bash
docker buildx build --platform linux/amd64 -f docker/frontend/Dockerfile -t jmartincufre/socialpredict-frontend:local --load .
```

## Run locally (smoke test)
1. Create a network:
```bash
docker network create spnet
```
2. Start Postgres (example):
```bash
docker run -d --name spdb --network spnet -e POSTGRES_USER=sp -e POSTGRES_PASSWORD=sp -e POSTGRES_DB=socialpredict postgres:16-alpine
```
3. Start backend (set required envs):
```bash
docker run -d --name spbackend --network spnet -p 8080:8080 \
  -e DB_HOST=spdb -e POSTGRES_USER=sp -e POSTGRES_PASSWORD=sp -e POSTGRES_DB=socialpredict \
  -e BACKEND_PORT=8080 -e JWT_SIGNING_KEY=yourkey -e ADMIN_PASSWORD=YourAdminPass \
  -e CORS_ALLOW_ORIGINS=http://localhost:8081 -e CORS_ALLOW_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS \
  -e CORS_ALLOW_HEADERS=Content-Type,Authorization -e CORS_ALLOW_CREDENTIALS=true \
  jmartincufre/socialpredict-backend:local
```
4. Start frontend (serves built assets and renders runtime env):
```bash
docker run -d --name spfrontend --network spnet -p 8081:80 \
  -e DOMAIN_URL=http://localhost:8081 -e API_URL=http://localhost:8080 \
  jmartincufre/socialpredict-frontend:local
```

5. Verify:
```bash
curl http://localhost:8080/v0/setup  # should return 200 JSON
curl http://localhost:8081/env-config.js  # should show window.__ENV__ with API_URL/DOMAIN_URL
```

## CI / Deployment notes
- Workflow triggers only on branch devops-docker; it logs into Docker Hub using secrets DOCKERHUB_USERNAME and DOCKERHUB_TOKEN and pushes images jmartincufre/socialpredict-backend and jmartincufre/socialpredict-frontend.
- Tags: latest and branch-specific tag (devops-docker-SHORTSHA).

## Environment variables (summary)
- Backend: DB_HOST, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB, DB_PORT (optional), BACKEND_PORT (default 8080), JWT_SIGNING_KEY, ADMIN_PASSWORD (for initial seed).
- Backend CORS (runtime): CORS_ENABLED (default true), CORS_ALLOW_ORIGINS (comma list, default *), CORS_ALLOW_METHODS (comma list, default GET,POST,PUT,PATCH,DELETE,OPTIONS), CORS_ALLOW_HEADERS (comma list, default Content-Type,Authorization), CORS_EXPOSE_HEADERS (optional), CORS_ALLOW_CREDENTIALS (default false), CORS_MAX_AGE (seconds, default 600).
- Frontend: DOMAIN_URL, API_URL (rendered at runtime to /env-config.js).
## Next steps and recommendations
- Add minimal health/ready probe endpoints (if not already) for orchestration.
- Replace any remaining bind-mounts in compose files with image-based deployment (separate PR).
- Add GitHub secrets and test workflow on branch `devops-docker`.
- Merge the new devops pipeline to main.

## Files touched
- [`docker/backend/Dockerfile`](docker/backend/Dockerfile:1)
- [`docker/frontend/Dockerfile`](docker/frontend/Dockerfile:1)
- [`docker/frontend/entrypoint.sh`](docker/frontend/entrypoint.sh:1)
- [`docker/frontend/nginx.conf`](docker/frontend/nginx.conf:1)
- [`frontend/public/env-config.js.template`](frontend/public/env-config.js.template:1)
- [`frontend/index.html`](frontend/index.html:29)
- [`backend/util/getenv.go`](backend/util/getenv.go:1)
- [`backend/util/postgres.go`](backend/util/postgres.go:1)
- [`.github/workflows/dockerhub-images.yml`](.github/workflows/dockerhub-images.yml:1)
