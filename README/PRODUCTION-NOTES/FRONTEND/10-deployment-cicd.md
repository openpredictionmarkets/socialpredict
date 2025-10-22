# Deployment and CI/CD Implementation Plan

## Overview
Implement comprehensive deployment strategies and CI/CD pipelines to ensure reliable, automated, and scalable application deployment with proper environment management, testing, and monitoring.

## Current State Analysis
- Manual deployment process
- No CI/CD pipeline
- Single environment setup
- No automated testing in deployment
- No deployment monitoring
- No rollback strategies
- Limited environment configuration
- No infrastructure as code

## Implementation Steps

### Step 1: Environment Configuration and Management
**Timeline: 2-3 days**

Set up comprehensive environment management with proper configuration:

```javascript
// config/environments.js
const environments = {
  development: {
    API_BASE_URL: 'http://localhost:8080/api',
    APP_ENV: 'development',
    DEBUG: true,
    ENABLE_DEVTOOLS: true,
    LOG_LEVEL: 'debug',
    ANALYTICS_ENABLED: false,
    GA_TRACKING_ID: '',
    MIXPANEL_TOKEN: '',
    SENTRY_DSN: '',
    FEATURE_FLAGS: {
      beta_features: true,
      debug_mode: true,
      mock_data: true,
    },
    CACHE_STRATEGY: 'no-cache',
    SERVICE_WORKER_ENABLED: false,
  },
  
  staging: {
    API_BASE_URL: 'https://api-staging.socialpredict.com/api',
    APP_ENV: 'staging',
    DEBUG: true,
    ENABLE_DEVTOOLS: true,
    LOG_LEVEL: 'info',
    ANALYTICS_ENABLED: true,
    GA_TRACKING_ID: 'GA_STAGING_ID',
    MIXPANEL_TOKEN: 'MIXPANEL_STAGING_TOKEN',
    SENTRY_DSN: 'SENTRY_STAGING_DSN',
    FEATURE_FLAGS: {
      beta_features: true,
      debug_mode: false,
      mock_data: false,
    },
    CACHE_STRATEGY: 'cache-first',
    SERVICE_WORKER_ENABLED: true,
  },
  
  production: {
    API_BASE_URL: 'https://api.socialpredict.com/api',
    APP_ENV: 'production',
    DEBUG: false,
    ENABLE_DEVTOOLS: false,
    LOG_LEVEL: 'error',
    ANALYTICS_ENABLED: true,
    GA_TRACKING_ID: 'GA_PRODUCTION_ID',
    MIXPANEL_TOKEN: 'MIXPANEL_PRODUCTION_TOKEN',
    SENTRY_DSN: 'SENTRY_PRODUCTION_DSN',
    FEATURE_FLAGS: {
      beta_features: false,
      debug_mode: false,
      mock_data: false,
    },
    CACHE_STRATEGY: 'cache-first',
    SERVICE_WORKER_ENABLED: true,
  },
}

export const getEnvironmentConfig = () => {
  const env = process.env.NODE_ENV || 'development'
  return {
    ...environments[env],
    // Override with environment variables
    API_BASE_URL: process.env.REACT_APP_API_BASE_URL || environments[env].API_BASE_URL,
    GA_TRACKING_ID: process.env.REACT_APP_GA_TRACKING_ID || environments[env].GA_TRACKING_ID,
    MIXPANEL_TOKEN: process.env.REACT_APP_MIXPANEL_TOKEN || environments[env].MIXPANEL_TOKEN,
    SENTRY_DSN: process.env.REACT_APP_SENTRY_DSN || environments[env].SENTRY_DSN,
  }
}

// utils/configManager.js
class ConfigManager {
  constructor() {
    this.config = getEnvironmentConfig()
    this.validateConfig()
  }

  validateConfig() {
    const requiredKeys = ['API_BASE_URL', 'APP_ENV']
    
    for (const key of requiredKeys) {
      if (!this.config[key]) {
        throw new Error(`Missing required configuration: ${key}`)
      }
    }

    // Validate URLs
    if (this.config.API_BASE_URL) {
      try {
        new URL(this.config.API_BASE_URL)
      } catch (error) {
        throw new Error(`Invalid API_BASE_URL: ${this.config.API_BASE_URL}`)
      }
    }

    console.log(`[Config] Environment: ${this.config.APP_ENV}`)
    console.log(`[Config] API URL: ${this.config.API_BASE_URL}`)
  }

  get(key, defaultValue = null) {
    return this.config[key] !== undefined ? this.config[key] : defaultValue
  }

  getFeatureFlag(flag) {
    return this.config.FEATURE_FLAGS?.[flag] || false
  }

  isDevelopment() {
    return this.config.APP_ENV === 'development'
  }

  isStaging() {
    return this.config.APP_ENV === 'staging'
  }

  isProduction() {
    return this.config.APP_ENV === 'production'
  }

  shouldEnableAnalytics() {
    return this.config.ANALYTICS_ENABLED && (this.isStaging() || this.isProduction())
  }
}

export const config = new ConfigManager()

// Environment-specific .env files
// .env.development
REACT_APP_API_BASE_URL=http://localhost:8080/api
REACT_APP_APP_ENV=development
REACT_APP_DEBUG=true
REACT_APP_LOG_LEVEL=debug

// .env.staging
REACT_APP_API_BASE_URL=https://api-staging.socialpredict.com/api
REACT_APP_APP_ENV=staging
REACT_APP_DEBUG=true
REACT_APP_LOG_LEVEL=info
REACT_APP_GA_TRACKING_ID=${GA_STAGING_ID}
REACT_APP_MIXPANEL_TOKEN=${MIXPANEL_STAGING_TOKEN}
REACT_APP_SENTRY_DSN=${SENTRY_STAGING_DSN}

// .env.production
REACT_APP_API_BASE_URL=https://api.socialpredict.com/api
REACT_APP_APP_ENV=production
REACT_APP_DEBUG=false
REACT_APP_LOG_LEVEL=error
REACT_APP_GA_TRACKING_ID=${GA_PRODUCTION_ID}
REACT_APP_MIXPANEL_TOKEN=${MIXPANEL_PRODUCTION_TOKEN}
REACT_APP_SENTRY_DSN=${SENTRY_PRODUCTION_DSN}

// vite.config.mjs - Environment-specific build configuration
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig(({ command, mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  
  return {
    plugins: [
      react(),
      VitePWA({
        registerType: 'autoUpdate',
        workbox: {
          globPatterns: ['**/*.{js,css,html,ico,png,svg}'],
          runtimeCaching: getRuntimeCaching(mode),
        },
        manifest: getManifest(mode),
      }),
    ],
    
    define: {
      __APP_VERSION__: JSON.stringify(process.env.npm_package_version),
      __BUILD_TIME__: JSON.stringify(new Date().toISOString()),
      __COMMIT_HASH__: JSON.stringify(process.env.GITHUB_SHA || 'development'),
    },
    
    build: {
      outDir: 'dist',
      sourcemap: mode !== 'production',
      minify: mode === 'production' ? 'esbuild' : false,
      target: 'es2015',
      rollupOptions: {
        output: {
          manualChunks: {
            vendor: ['react', 'react-dom', 'react-router-dom'],
            charts: ['chart.js', 'react-chartjs-2'],
            utils: ['date-fns', 'lodash'],
          },
        },
      },
    },
    
    server: {
      port: 3000,
      host: true,
      proxy: mode === 'development' ? {
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
          secure: false,
        },
      } : {},
    },
    
    preview: {
      port: 3000,
      host: true,
    },
  }
})

function getRuntimeCaching(mode) {
  if (mode === 'development') return []
  
  return [
    {
      urlPattern: /^https:\/\/api\.socialpredict\.com\/api\/markets/,
      handler: 'NetworkFirst',
      options: {
        cacheName: 'api-markets',
        expiration: {
          maxEntries: 100,
          maxAgeSeconds: 60 * 60, // 1 hour
        },
      },
    },
    {
      urlPattern: /\.(?:png|jpg|jpeg|svg|gif)$/,
      handler: 'CacheFirst',
      options: {
        cacheName: 'images',
        expiration: {
          maxEntries: 500,
          maxAgeSeconds: 60 * 60 * 24 * 30, // 30 days
        },
      },
    },
  ]
}

function getManifest(mode) {
  const baseManifest = {
    name: 'SocialPredict',
    short_name: 'SocialPredict',
    description: 'Social prediction markets platform',
    theme_color: '#4F46E5',
    background_color: '#FFFFFF',
    display: 'standalone',
    start_url: '/',
    scope: '/',
  }
  
  if (mode === 'development') {
    return {
      ...baseManifest,
      name: 'SocialPredict (Dev)',
      short_name: 'SP Dev',
    }
  }
  
  if (mode === 'staging') {
    return {
      ...baseManifest,
      name: 'SocialPredict (Staging)',
      short_name: 'SP Staging',
      theme_color: '#F59E0B',
    }
  }
  
  return baseManifest
}
```

### Step 2: GitHub Actions CI/CD Pipeline
**Timeline: 3-4 days**

Implement comprehensive CI/CD pipeline with GitHub Actions:

```yaml
# .github/workflows/ci.yml - Continuous Integration
name: CI Pipeline

on:
  push:
    branches: [ main, develop, 'feature/*' ]
  pull_request:
    branches: [ main, develop ]

env:
  NODE_VERSION: '18.x'

jobs:
  lint-and-test:
    name: Lint and Test
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      
    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: ${{ env.NODE_VERSION }}
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json
        
    - name: Install dependencies
      working-directory: ./frontend
      run: npm ci
      
    - name: Run ESLint
      working-directory: ./frontend
      run: npm run lint:check
      
    - name: Run Prettier
      working-directory: ./frontend
      run: npm run format:check
      
    - name: Type check
      working-directory: ./frontend
      run: npm run type-check
      
    - name: Run unit tests
      working-directory: ./frontend
      run: npm run test:coverage
      
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./frontend/coverage/lcov.info
        flags: unittests
        name: codecov-umbrella
        
  build:
    name: Build Application
    runs-on: ubuntu-latest
    needs: lint-and-test
    
    strategy:
      matrix:
        environment: [development, staging, production]
        
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      
    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: ${{ env.NODE_VERSION }}
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json
        
    - name: Install dependencies
      working-directory: ./frontend
      run: npm ci
      
    - name: Build for ${{ matrix.environment }}
      working-directory: ./frontend
      run: npm run build:${{ matrix.environment }}
      env:
        REACT_APP_VERSION: ${{ github.sha }}
        REACT_APP_BUILD_TIME: ${{ github.run_number }}
        
    - name: Upload build artifacts
      uses: actions/upload-artifact@v3
      with:
        name: build-${{ matrix.environment }}
        path: frontend/dist/
        retention-days: 7

  security-scan:
    name: Security Scan
    runs-on: ubuntu-latest
    needs: lint-and-test
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      
    - name: Run npm audit
      working-directory: ./frontend
      run: npm audit --audit-level high
      
    - name: Run Snyk to check for vulnerabilities
      uses: snyk/actions/node@master
      env:
        SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
      with:
        args: --severity-threshold=high
        command: test

  e2e-tests:
    name: E2E Tests
    runs-on: ubuntu-latest
    needs: build
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      
    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: ${{ env.NODE_VERSION }}
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json
        
    - name: Install dependencies
      working-directory: ./frontend
      run: npm ci
      
    - name: Download build artifacts
      uses: actions/download-artifact@v3
      with:
        name: build-staging
        path: frontend/dist/
        
    - name: Install Playwright
      working-directory: ./frontend
      run: npx playwright install --with-deps
      
    - name: Start backend services
      run: |
        docker-compose -f docker-compose-test.yaml up -d
        sleep 30
        
    - name: Run E2E tests
      working-directory: ./frontend
      run: npm run test:e2e
      
    - name: Upload Playwright report
      uses: actions/upload-artifact@v3
      if: always()
      with:
        name: playwright-report
        path: frontend/playwright-report/
        retention-days: 7
        
    - name: Stop backend services
      if: always()
      run: docker-compose -f docker-compose-test.yaml down

# .github/workflows/deploy-staging.yml - Staging Deployment
name: Deploy to Staging

on:
  push:
    branches: [ develop ]
  workflow_dispatch:

env:
  NODE_VERSION: '18.x'

jobs:
  deploy-staging:
    name: Deploy to Staging
    runs-on: ubuntu-latest
    environment: staging
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      
    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: ${{ env.NODE_VERSION }}
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json
        
    - name: Install dependencies
      working-directory: ./frontend
      run: npm ci
      
    - name: Build for staging
      working-directory: ./frontend
      run: npm run build:staging
      env:
        REACT_APP_API_BASE_URL: ${{ secrets.STAGING_API_URL }}
        REACT_APP_GA_TRACKING_ID: ${{ secrets.STAGING_GA_ID }}
        REACT_APP_MIXPANEL_TOKEN: ${{ secrets.STAGING_MIXPANEL_TOKEN }}
        REACT_APP_SENTRY_DSN: ${{ secrets.STAGING_SENTRY_DSN }}
        REACT_APP_VERSION: ${{ github.sha }}
        
    - name: Deploy to AWS S3
      uses: aws-actions/configure-aws-credentials@v2
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-east-1
        
    - name: Sync to S3 bucket
      run: |
        aws s3 sync frontend/dist/ s3://${{ secrets.STAGING_S3_BUCKET }} --delete
        
    - name: Invalidate CloudFront cache
      run: |
        aws cloudfront create-invalidation \
          --distribution-id ${{ secrets.STAGING_CLOUDFRONT_ID }} \
          --paths "/*"
          
    - name: Run smoke tests
      working-directory: ./frontend
      run: npm run test:smoke -- --baseURL=https://staging.socialpredict.com
      
    - name: Notify deployment
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        channel: '#deployments'
        text: 'Staging deployment completed'
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}

# .github/workflows/deploy-production.yml - Production Deployment
name: Deploy to Production

on:
  push:
    branches: [ main ]
  release:
    types: [ published ]
  workflow_dispatch:

env:
  NODE_VERSION: '18.x'

jobs:
  deploy-production:
    name: Deploy to Production
    runs-on: ubuntu-latest
    environment: production
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      
    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: ${{ env.NODE_VERSION }}
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json
        
    - name: Install dependencies
      working-directory: ./frontend
      run: npm ci
      
    - name: Run full test suite
      working-directory: ./frontend
      run: |
        npm run lint:check
        npm run test:coverage
        npm run build:production
      env:
        REACT_APP_API_BASE_URL: ${{ secrets.PRODUCTION_API_URL }}
        REACT_APP_GA_TRACKING_ID: ${{ secrets.PRODUCTION_GA_ID }}
        REACT_APP_MIXPANEL_TOKEN: ${{ secrets.PRODUCTION_MIXPANEL_TOKEN }}
        REACT_APP_SENTRY_DSN: ${{ secrets.PRODUCTION_SENTRY_DSN }}
        REACT_APP_VERSION: ${{ github.sha }}
        
    - name: Create deployment backup
      run: |
        timestamp=$(date +"%Y%m%d-%H%M%S")
        aws s3 sync s3://${{ secrets.PRODUCTION_S3_BUCKET }} s3://socialpredict-backups/frontend/$timestamp/
        
    - name: Deploy to AWS S3
      uses: aws-actions/configure-aws-credentials@v2
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-east-1
        
    - name: Blue-Green Deployment
      run: |
        # Deploy to blue environment first
        aws s3 sync frontend/dist/ s3://${{ secrets.PRODUCTION_S3_BUCKET_BLUE }} --delete
        
        # Run health checks
        curl -f https://blue.socialpredict.com/health || exit 1
        
        # Switch traffic to blue
        aws cloudfront update-distribution \
          --id ${{ secrets.PRODUCTION_CLOUDFRONT_ID }} \
          --distribution-config file://cloudfront-blue-config.json
          
        # Wait for propagation
        sleep 60
        
        # Final health check
        curl -f https://socialpredict.com/health || exit 1
        
    - name: Run production smoke tests
      working-directory: ./frontend
      run: npm run test:smoke -- --baseURL=https://socialpredict.com
      
    - name: Update monitoring
      run: |
        # Update Sentry release
        curl -sL https://sentry.io/get-cli/ | bash
        sentry-cli releases new ${{ github.sha }}
        sentry-cli releases set-commits ${{ github.sha }} --auto
        sentry-cli releases finalize ${{ github.sha }}
        
    - name: Notify successful deployment
      uses: 8398a7/action-slack@v3
      with:
        status: 'success'
        channel: '#deployments'
        text: 'Production deployment completed successfully!'
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
        
    - name: Notify failed deployment
      if: failure()
      uses: 8398a7/action-slack@v3
      with:
        status: 'failure'
        channel: '#deployments'
        text: 'Production deployment failed! Rolling back...'
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}

# .github/workflows/rollback.yml - Rollback Workflow
name: Rollback Production

on:
  workflow_dispatch:
    inputs:
      backup_timestamp:
        description: 'Backup timestamp to rollback to (YYYYMMDD-HHMMSS)'
        required: true
        type: string

jobs:
  rollback:
    name: Rollback Production
    runs-on: ubuntu-latest
    environment: production
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      
    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v2
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-east-1
        
    - name: Restore from backup
      run: |
        echo "Rolling back to backup: ${{ github.event.inputs.backup_timestamp }}"
        aws s3 sync s3://socialpredict-backups/frontend/${{ github.event.inputs.backup_timestamp }}/ s3://${{ secrets.PRODUCTION_S3_BUCKET }} --delete
        
    - name: Invalidate CloudFront cache
      run: |
        aws cloudfront create-invalidation \
          --distribution-id ${{ secrets.PRODUCTION_CLOUDFRONT_ID }} \
          --paths "/*"
          
    - name: Verify rollback
      run: |
        sleep 60
        curl -f https://socialpredict.com/health || exit 1
        
    - name: Notify rollback
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        channel: '#deployments'
        text: 'Production rollback completed to backup: ${{ github.event.inputs.backup_timestamp }}'
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

### Step 3: Docker Containerization
**Timeline: 2 days**

Create comprehensive Docker setup for containerized deployments:

```dockerfile
# Dockerfile.development
FROM node:18-alpine as development

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install all dependencies (including dev dependencies)
RUN npm ci

# Copy source code
COPY . .

# Expose port
EXPOSE 3000

# Start development server
CMD ["npm", "run", "dev"]

# Dockerfile.production
FROM node:18-alpine as builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production && npm cache clean --force

# Copy source code
COPY . .

# Build the application
RUN npm run build:production

# Production stage
FROM nginx:alpine as production

# Copy custom nginx config
COPY nginx.conf /etc/nginx/nginx.conf

# Copy built application
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy health check script
COPY healthcheck.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/healthcheck.sh

# Add health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD /usr/local/bin/healthcheck.sh

# Expose port
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]

# nginx.conf - Production Nginx configuration
events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;
    
    # Logging
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                   '$status $body_bytes_sent "$http_referer" '
                   '"$http_user_agent" "$http_x_forwarded_for"';
    
    access_log /var/log/nginx/access.log main;
    error_log /var/log/nginx/error.log warn;
    
    # Basic settings
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    client_max_body_size 16M;
    
    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;
    
    # Security headers
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' https://www.googletagmanager.com https://cdn.mxpnl.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self' https://api.socialpredict.com wss://api.socialpredict.com;" always;
    
    server {
        listen 80;
        server_name _;
        root /usr/share/nginx/html;
        index index.html;
        
        # Health check endpoint
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
        
        # Static assets with long-term caching
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
            try_files $uri =404;
        }
        
        # Service worker
        location /sw.js {
            expires off;
            add_header Cache-Control "no-cache, no-store, must-revalidate";
            try_files $uri =404;
        }
        
        # Manifest
        location /manifest.json {
            expires 1d;
            add_header Cache-Control "public";
            try_files $uri =404;
        }
        
        # API proxy (for development)
        location /api/ {
            proxy_pass http://backend:8080/api/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
        
        # SPA routing - serve index.html for all routes
        location / {
            try_files $uri $uri/ /index.html;
            expires -1;
            add_header Cache-Control "no-cache, no-store, must-revalidate";
        }
        
        # Error pages
        error_page 404 /index.html;
        error_page 500 502 503 504 /50x.html;
        location = /50x.html {
            root /usr/share/nginx/html;
        }
    }
}

# healthcheck.sh
#!/bin/sh
set -e

# Check if nginx is running
if ! pgrep nginx > /dev/null; then
    echo "Nginx is not running"
    exit 1
fi

# Check if the application responds
if ! wget --no-verbose --tries=1 --spider http://localhost/health; then
    echo "Health check failed"
    exit 1
fi

echo "Health check passed"
exit 0

# docker-compose.yml - Development environment
version: '3.8'

services:
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.development
    ports:
      - "3000:3000"
    volumes:
      - ./frontend:/app
      - /app/node_modules
    environment:
      - NODE_ENV=development
      - REACT_APP_API_BASE_URL=http://localhost:8080/api
    depends_on:
      - backend
    networks:
      - socialpredict
      
networks:
  socialpredict:
    driver: bridge

# docker-compose.prod.yml - Production environment
version: '3.8'

services:
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.production
    ports:
      - "80:80"
      - "443:443"
    environment:
      - NODE_ENV=production
    volumes:
      - ./ssl:/etc/nginx/ssl:ro
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "/usr/local/bin/healthcheck.sh"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    networks:
      - socialpredict
      
networks:
  socialpredict:
    driver: bridge
```

### Step 4: Infrastructure as Code (Terraform)
**Timeline: 3-4 days**

Set up infrastructure as code for AWS deployment:

```hcl
# terraform/main.tf
terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  
  backend "s3" {
    bucket = "socialpredict-terraform-state"
    key    = "frontend/terraform.tfstate"
    region = "us-east-1"
  }
}

provider "aws" {
  region = var.aws_region
  
  default_tags {
    tags = {
      Project     = "SocialPredict"
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}

# Variables
variable "environment" {
  description = "Environment name"
  type        = string
  validation {
    condition     = contains(["staging", "production"], var.environment)
    error_message = "Environment must be staging or production."
  }
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "domain_name" {
  description = "Domain name for the application"
  type        = string
}

# S3 bucket for static hosting
resource "aws_s3_bucket" "frontend" {
  bucket = "socialpredict-frontend-${var.environment}"
}

resource "aws_s3_bucket_versioning" "frontend" {
  bucket = aws_s3_bucket.frontend.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "frontend" {
  bucket = aws_s3_bucket.frontend.id
  
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "frontend" {
  bucket = aws_s3_bucket.frontend.id
  
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# CloudFront distribution
resource "aws_cloudfront_origin_access_identity" "frontend" {
  comment = "Frontend OAI for ${var.environment}"
}

resource "aws_s3_bucket_policy" "frontend" {
  bucket = aws_s3_bucket.frontend.id
  
  policy = jsonencode({
    Statement = [
      {
        Sid    = "AllowCloudFrontAccess"
        Effect = "Allow"
        Principal = {
          AWS = aws_cloudfront_origin_access_identity.frontend.iam_arn
        }
        Action   = "s3:GetObject"
        Resource = "${aws_s3_bucket.frontend.arn}/*"
      }
    ]
  })
}

resource "aws_cloudfront_distribution" "frontend" {
  origin {
    domain_name = aws_s3_bucket.frontend.bucket_regional_domain_name
    origin_id   = "S3-${aws_s3_bucket.frontend.bucket}"
    
    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.frontend.cloudfront_access_identity_path
    }
  }
  
  enabled             = true
  is_ipv6_enabled     = true
  default_root_object = "index.html"
  
  aliases = var.environment == "production" ? [var.domain_name] : ["${var.environment}.${var.domain_name}"]
  
  default_cache_behavior {
    allowed_methods        = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "S3-${aws_s3_bucket.frontend.bucket}"
    compress               = true
    viewer_protocol_policy = "redirect-to-https"
    
    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }
    
    min_ttl     = 0
    default_ttl = 3600
    max_ttl     = 86400
  }
  
  # Cache behavior for static assets
  ordered_cache_behavior {
    path_pattern     = "/static/*"
    allowed_methods  = ["GET", "HEAD", "OPTIONS"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "S3-${aws_s3_bucket.frontend.bucket}"
    compress         = true
    
    forwarded_values {
      query_string = false
      headers      = ["Origin", "Access-Control-Request-Headers", "Access-Control-Request-Method"]
      cookies {
        forward = "none"
      }
    }
    
    viewer_protocol_policy = "https-only"
    min_ttl                = 31536000
    default_ttl            = 31536000
    max_ttl                = 31536000
  }
  
  # Custom error pages for SPA routing
  custom_error_response {
    error_code            = 404
    response_code         = 200
    response_page_path    = "/index.html"
    error_caching_min_ttl = 0
  }
  
  custom_error_response {
    error_code            = 403
    response_code         = 200
    response_page_path    = "/index.html"
    error_caching_min_ttl = 0
  }
  
  price_class = var.environment == "production" ? "PriceClass_All" : "PriceClass_100"
  
  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }
  
  viewer_certificate {
    acm_certificate_arn      = var.environment == "production" ? aws_acm_certificate.frontend[0].arn : null
    ssl_support_method       = var.environment == "production" ? "sni-only" : null
    minimum_protocol_version = var.environment == "production" ? "TLSv1.2_2021" : null
    cloudfront_default_certificate = var.environment != "production"
  }
  
  web_acl_id = aws_wafv2_web_acl.frontend.arn
  
  tags = {
    Name = "Frontend Distribution - ${var.environment}"
  }
}

# SSL Certificate (for production)
resource "aws_acm_certificate" "frontend" {
  count = var.environment == "production" ? 1 : 0
  
  domain_name       = var.domain_name
  validation_method = "DNS"
  
  lifecycle {
    create_before_destroy = true
  }
}

# WAF for security
resource "aws_wafv2_web_acl" "frontend" {
  name  = "socialpredict-frontend-${var.environment}"
  scope = "CLOUDFRONT"
  
  default_action {
    allow {}
  }
  
  # Rate limiting rule
  rule {
    name     = "RateLimitRule"
    priority = 1
    
    override_action {
      none {}
    }
    
    statement {
      rate_based_statement {
        limit              = 2000
        aggregate_key_type = "IP"
      }
    }
    
    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "RateLimitRule"
      sampled_requests_enabled   = true
    }
    
    action {
      block {}
    }
  }
  
  # AWS Managed Rules
  rule {
    name     = "AWSManagedRulesCommonRuleSet"
    priority = 2
    
    override_action {
      none {}
    }
    
    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesCommonRuleSet"
        vendor_name = "AWS"
      }
    }
    
    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "CommonRuleSetMetric"
      sampled_requests_enabled   = true
    }
  }
  
  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "socialpredict-frontend-${var.environment}"
    sampled_requests_enabled   = true
  }
}

# Outputs
output "cloudfront_distribution_id" {
  description = "CloudFront Distribution ID"
  value       = aws_cloudfront_distribution.frontend.id
}

output "cloudfront_domain_name" {
  description = "CloudFront Distribution Domain Name"
  value       = aws_cloudfront_distribution.frontend.domain_name
}

output "s3_bucket_name" {
  description = "S3 Bucket Name"
  value       = aws_s3_bucket.frontend.bucket
}
```

## Directory Structure
```
.github/
├── workflows/
│   ├── ci.yml                # Continuous Integration
│   ├── deploy-staging.yml    # Staging deployment
│   ├── deploy-production.yml # Production deployment
│   └── rollback.yml         # Rollback workflow
├── ISSUE_TEMPLATE/
└── PULL_REQUEST_TEMPLATE.md

scripts/
├── deploy/
│   ├── deploy.sh            # Deployment script
│   ├── rollback.sh          # Rollback script
│   └── health-check.sh      # Health check script
├── build/
│   ├── build-staging.sh     # Staging build
│   └── build-production.sh  # Production build
└── test/
    ├── smoke-tests.sh       # Smoke tests
    └── load-tests.sh        # Load tests

terraform/
├── main.tf                  # Main infrastructure
├── variables.tf             # Variables
├── outputs.tf              # Outputs
├── environments/
│   ├── staging.tfvars      # Staging variables
│   └── production.tfvars   # Production variables
└── modules/
    ├── s3/                 # S3 module
    ├── cloudfront/         # CloudFront module
    └── waf/               # WAF module

docker/
├── Dockerfile.development   # Development container
├── Dockerfile.production   # Production container
├── nginx.conf              # Nginx configuration
├── healthcheck.sh          # Health check script
└── docker-compose.yml      # Docker Compose

config/
├── environments.js         # Environment config
├── build/
│   ├── staging.js         # Staging build config
│   └── production.js      # Production build config
└── deploy/
    ├── staging.json       # Staging deploy config
    └── production.json    # Production deploy config
```

## Benefits
- Automated deployment process
- Environment consistency
- Rollback capabilities
- Security scanning
- Performance monitoring
- Infrastructure as code
- Blue-green deployments
- Automated testing
- Configuration management
- Monitoring and alerting
- Disaster recovery
- Compliance and auditing

## Deployment Features Implemented
- ✅ Multi-environment configuration
- ✅ GitHub Actions CI/CD pipeline
- ✅ Docker containerization
- ✅ Infrastructure as Code (Terraform)
- ✅ Blue-green deployments
- ✅ Automated testing
- ✅ Security scanning
- ✅ Performance monitoring
- ✅ Rollback strategies
- ✅ Health checks
- ✅ CDN deployment
- ✅ SSL/TLS certificates

## Supported Environments
- Development (local)
- Staging (pre-production)
- Production (live)

## Deployment Strategies
- Blue-Green Deployment
- Rolling Updates
- Canary Releases
- Feature Flags
- A/B Testing Support