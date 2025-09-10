FROM node:22-alpine AS build
RUN apk add --no-cache python3 make g++ build-base
WORKDIR /app
# Copy only package files first for better caching
COPY frontend/package.json frontend/package-lock.json ./frontend/
WORKDIR /app/frontend
RUN npm ci --production=false
# Copy the rest of the frontend sources and build
COPY frontend/ .
RUN npm run build

FROM nginx:alpine
# Copy built site
COPY --from=build /app/frontend/build /usr/share/nginx/html
# Copy runtime template and entrypoint
COPY frontend/public/env-config.js.template /usr/share/nginx/html/env-config.js.template
COPY docker/frontend/nginx.conf /etc/nginx/conf.d/default.conf
COPY docker/frontend/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Install envplate
RUN apk add --no-cache wget tar && \
    wget -q https://github.com/kreuzwerker/envplate/releases/download/v1.0.3/envplate_1.0.3_$(uname -s)_$(uname -m).tar.gz -O - | tar xz && \
    mv envplate /usr/local/bin/ep && chmod +x /usr/local/bin/ep && \
    rm -rf /var/cache/apk/*

EXPOSE 80
ENTRYPOINT ["/entrypoint.sh"]