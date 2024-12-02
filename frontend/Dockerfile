# syntax=docker/dockerfile:1.3-labs
FROM denoland/deno:debian

WORKDIR /app

COPY package.json package-lock.json ./
RUN deno i
COPY . .

EXPOSE 5173
CMD ["deno", "run", "start"]
