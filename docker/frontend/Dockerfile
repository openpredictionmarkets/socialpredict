FROM node:22.20.0-alpine3.22 AS builder

WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./frontend/
WORKDIR /app/frontend
RUN npm ci
COPY frontend/ .
RUN npm run build

FROM nginx:1.29.1-alpine3.22

COPY --from=builder /app/frontend/build /usr/share/nginx/html
COPY frontend/public/env-config.js.template /usr/share/nginx/html/env-config.js.template
COPY docker/frontend/nginx.conf /etc/nginx/conf.d/default.conf
COPY docker/frontend/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
RUN chown -R nginx:nginx /usr/share/nginx/html
EXPOSE 80
ENTRYPOINT ["/entrypoint.sh"]
