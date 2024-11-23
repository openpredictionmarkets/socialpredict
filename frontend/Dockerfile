# syntax=docker/dockerfile:1.3-labs
FROM deno:lastest

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm install
COPY . .

EXPOSE 5173
CMD ["deno", "run", "start"]
