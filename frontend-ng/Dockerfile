# Multi-stage build to produce a runnable SvelteKit app
FROM node:lts-alpine AS builder
WORKDIR /app

# Install dependencies and build
COPY package.json pnpm-lock.yaml ./
RUN corepack enable \
  && corepack prepare pnpm@latest --activate \
  && pnpm install --frozen-lockfile
COPY . .
RUN pnpm build

FROM node:lts-alpine
WORKDIR /app

ENV NODE_ENV=production
ENV PORT=7222

# Bring over build artifacts and deps
COPY --from=builder /app/build ./build
COPY --from=builder /app/node_modules ./node_modules

EXPOSE 7222
CMD ["node", "./build"]
