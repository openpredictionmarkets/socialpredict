# Deployment & Infrastructure Implementation Plan

## Overview
Enhance the existing Docker setup and create production-ready deployment infrastructure with container optimization, orchestration, CI/CD pipelines, and monitoring integration.

## Current State Analysis
- Basic Dockerfile exists for backend
- Docker Compose configurations for dev and prod
- No container optimization
- Limited health checks
- No automated deployment pipeline
- Basic container orchestration

## Implementation Steps

### Step 1: Container Optimization
**Timeline: 2-3 days**

Optimize the existing Docker container for production:

```dockerfile
# Dockerfile.production
# Multi-stage build for smaller image size
FROM golang:1.23-alpine AS builder

# Install necessary packages for building
RUN apk add --no-cache git ca-certificates tzdata

# Create appuser for security
RUN adduser -D -g '' appuser

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo -o socialpredict ./main.go

# Final stage - minimal runtime image
FROM scratch

# Import certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Import user from builder stage
COPY --from=builder /etc/passwd /etc/passwd

# Copy binary from builder stage
COPY --from=builder /build/socialpredict /socialpredict

# Use non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/socialpredict", "health"]

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/socialpredict"]
```

**Container optimizations:**
- Multi-stage builds
- Minimal base images
- Security hardening
- Health checks
- Proper signal handling
- Resource limits

### Step 2: Kubernetes Deployment Manifests
**Timeline: 3-4 days**

Create comprehensive Kubernetes deployment configuration:

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: socialpredict
  labels:
    name: socialpredict

---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: socialpredict-config
  namespace: socialpredict
data:
  app.yaml: |
    server:
      port: 8080
      timeout: 30s
    database:
      max_open_conns: 25
      max_idle_conns: 10
    logging:
      level: "info"
      format: "json"

---
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: socialpredict-secrets
  namespace: socialpredict
type: Opaque
data:
  # Base64 encoded secrets
  postgres-password: <base64-encoded-password>
  jwt-secret: <base64-encoded-jwt-secret>

---
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: socialpredict-backend
  namespace: socialpredict
  labels:
    app: socialpredict-backend
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: socialpredict-backend
  template:
    metadata:
      labels:
        app: socialpredict-backend
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: socialpredict
        image: socialpredict:latest
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: CONFIG_PATH
          value: "/config/app.yaml"
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: socialpredict-secrets
              key: postgres-password
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: socialpredict-secrets
              key: jwt-secret
        volumeMounts:
        - name: config
          mountPath: /config
          readOnly: true
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        startupProbe:
          httpGet:
            path: /health/startup
            port: 8080
          failureThreshold: 30
          periodSeconds: 10
      volumes:
      - name: config
        configMap:
          name: socialpredict-config
      serviceAccountName: socialpredict-backend
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        fsGroup: 65534

---
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: socialpredict-backend-service
  namespace: socialpredict
  labels:
    app: socialpredict-backend
spec:
  selector:
    app: socialpredict-backend
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
  type: ClusterIP

---
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: socialpredict-ingress
  namespace: socialpredict
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - api.socialpredict.com
    secretName: socialpredict-tls
  rules:
  - host: api.socialpredict.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: socialpredict-backend-service
            port:
              number: 80
```

### Step 3: Helm Chart Creation
**Timeline: 2-3 days**

Create Helm chart for easier deployment management:

```yaml
# helm/socialpredict/Chart.yaml
apiVersion: v2
name: socialpredict
description: SocialPredict Backend Helm Chart
type: application
version: 0.1.0
appVersion: "1.0.0"

# helm/socialpredict/values.yaml
replicaCount: 3

image:
  repository: socialpredict/backend
  pullPolicy: IfNotPresent
  tag: ""

service:
  type: ClusterIP
  port: 80
  targetPort: 8080

ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: api.socialpredict.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: socialpredict-tls
      hosts:
        - api.socialpredict.com

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

database:
  host: postgresql
  port: 5432
  name: socialpredict
  auth:
    username: postgres
    existingSecret: socialpredict-secrets
    secretKey: postgres-password

config:
  logLevel: info
  serverTimeout: 30s
  jwtExpiry: 15m
```

### Step 4: CI/CD Pipeline Implementation
**Timeline: 3-4 days**

Create comprehensive CI/CD pipeline:

```yaml
# .github/workflows/deploy.yml
name: Build and Deploy

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.23

    - name: Run tests
      run: make test

    - name: Run security scan
      uses: securecodewarrior/github-action-gosec@master
      with:
        sarif_file: 'gosec.sarif'

  build:
    needs: test
    runs-on: ubuntu-latest
    outputs:
      image: ${{ steps.image.outputs.image }}
      digest: ${{ steps.build.outputs.digest }}
    steps:
    - name: Checkout
      uses: actions/checkout@v3

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2

    - name: Log in to Container Registry
      uses: docker/login-action@v2
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v4
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=sha,prefix={{branch}}-
          type=raw,value=latest,enable={{is_default_branch}}

    - name: Build and push
      id: build
      uses: docker/build-push-action@v4
      with:
        context: .
        file: ./Dockerfile.production
        platforms: linux/amd64,linux/arm64
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

    - name: Output image
      id: image
      run: |
        echo "image=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}" >> $GITHUB_OUTPUT

  deploy-staging:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    environment: staging
    steps:
    - name: Checkout
      uses: actions/checkout@v3

    - name: Configure kubectl
      uses: azure/k8s-set-context@v1
      with:
        method: kubeconfig
        kubeconfig: ${{ secrets.KUBE_CONFIG_STAGING }}

    - name: Deploy to staging
      run: |
        helm upgrade --install socialpredict-staging ./helm/socialpredict \
          --namespace socialpredict-staging \
          --create-namespace \
          --set image.tag=${{ github.sha }} \
          --set ingress.hosts[0].host=staging-api.socialpredict.com \
          --values ./helm/socialpredict/values-staging.yaml

    - name: Run health check
      run: |
        kubectl wait --for=condition=ready pod -l app=socialpredict-backend \
          -n socialpredict-staging --timeout=300s

        # Run basic health check
        kubectl run health-check --image=curlimages/curl --rm -i --restart=Never \
          -- curl -f http://socialpredict-backend-service.socialpredict-staging/health

  deploy-production:
    needs: [build, deploy-staging]
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    environment: production
    steps:
    - name: Checkout
      uses: actions/checkout@v3

    - name: Configure kubectl
      uses: azure/k8s-set-context@v1
      with:
        method: kubeconfig
        kubeconfig: ${{ secrets.KUBE_CONFIG_PRODUCTION }}

    - name: Deploy to production
      run: |
        helm upgrade --install socialpredict ./helm/socialpredict \
          --namespace socialpredict \
          --create-namespace \
          --set image.tag=${{ github.sha }} \
          --values ./helm/socialpredict/values-production.yaml

    - name: Verify deployment
      run: |
        kubectl rollout status deployment/socialpredict-backend \
          -n socialpredict --timeout=300s
```

### Step 5: Infrastructure as Code
**Timeline: 2-3 days**

Create Terraform configurations for cloud infrastructure:

```hcl
# terraform/main.tf
terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }
}

# terraform/cluster.tf
resource "kubernetes_namespace" "socialpredict" {
  metadata {
    name = "socialpredict"
    labels = {
      name = "socialpredict"
    }
  }
}

resource "kubernetes_secret" "socialpredict_secrets" {
  metadata {
    name      = "socialpredict-secrets"
    namespace = kubernetes_namespace.socialpredict.metadata[0].name
  }

  data = {
    postgres-password = var.postgres_password
    jwt-secret       = var.jwt_secret
  }

  type = "Opaque"
}

# terraform/monitoring.tf
resource "helm_release" "prometheus" {
  name       = "prometheus"
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  namespace  = "monitoring"
  create_namespace = true

  values = [
    file("${path.module}/monitoring-values.yaml")
  ]
}

# terraform/database.tf
resource "kubernetes_deployment" "postgresql" {
  metadata {
    name      = "postgresql"
    namespace = kubernetes_namespace.socialpredict.metadata[0].name
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "postgresql"
      }
    }

    template {
      metadata {
        labels = {
          app = "postgresql"
        }
      }

      spec {
        container {
          image = "postgres:13"
          name  = "postgresql"

          env {
            name  = "POSTGRES_DB"
            value = "socialpredict"
          }

          env {
            name = "POSTGRES_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.socialpredict_secrets.metadata[0].name
                key  = "postgres-password"
              }
            }
          }

          port {
            container_port = 5432
          }

          volume_mount {
            name       = "postgresql-storage"
            mount_path = "/var/lib/postgresql/data"
          }
        }

        volume {
          name = "postgresql-storage"
          persistent_volume_claim {
            claim_name = kubernetes_persistent_volume_claim.postgresql.metadata[0].name
          }
        }
      }
    }
  }
}
```

### Step 6: Monitoring and Observability Integration
**Timeline: 2 days**

Integrate with monitoring and observability stack:

```yaml
# k8s/monitoring/servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: socialpredict-backend
  namespace: socialpredict
  labels:
    app: socialpredict-backend
spec:
  selector:
    matchLabels:
      app: socialpredict-backend
  endpoints:
  - port: http
    path: /metrics
    interval: 30s

---
# k8s/monitoring/prometheusrule.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: socialpredict-alerts
  namespace: socialpredict
spec:
  groups:
  - name: socialpredict.rules
    rules:
    - alert: HighErrorRate
      expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.1
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: High error rate detected
        description: "Error rate is {{ $value | humanizePercentage }}"

    - alert: HighResponseTime
      expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: High response time detected
        description: "95th percentile response time is {{ $value }}s"

    - alert: PodCrashLooping
      expr: rate(kube_pod_container_status_restarts_total[15m]) > 0
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: Pod is crash looping
        description: "Pod {{ $labels.pod }} is crash looping"
```

### Step 7: Security and Compliance
**Timeline: 2-3 days**

Implement security best practices and compliance:

```yaml
# k8s/security/networkpolicy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: socialpredict-network-policy
  namespace: socialpredict
spec:
  podSelector:
    matchLabels:
      app: socialpredict-backend
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgresql
    ports:
    - protocol: TCP
      port: 5432
  - to: []
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53

---
# k8s/security/podsecuritypolicy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: socialpredict-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
```

## Directory Structure
```
deployment/
├── docker/
│   ├── Dockerfile.production    # Optimized production Dockerfile
│   ├── Dockerfile.development   # Development Dockerfile
│   └── docker-compose.prod.yml  # Production docker-compose
├── k8s/
│   ├── base/                    # Base Kubernetes manifests
│   ├── overlays/               # Kustomize overlays
│   ├── monitoring/             # Monitoring configurations
│   └── security/               # Security policies
├── helm/
│   └── socialpredict/          # Helm chart
│       ├── Chart.yaml
│       ├── values.yaml
│       ├── values-staging.yaml
│       ├── values-production.yaml
│       └── templates/
├── terraform/
│   ├── main.tf                 # Main Terraform configuration
│   ├── variables.tf            # Input variables
│   ├── outputs.tf              # Output values
│   └── modules/                # Terraform modules
└── scripts/
    ├── deploy.sh               # Deployment scripts
    ├── health-check.sh         # Health check scripts
    └── rollback.sh             # Rollback scripts
```

## Deployment Environments

### Development
- Single replica
- Reduced resource limits
- Debug logging enabled
- Hot-reload capabilities

### Staging
- Production-like configuration
- Automated testing integration
- Performance testing
- Security scanning

### Production
- High availability (3+ replicas)
- Resource optimization
- Monitoring and alerting
- Backup and disaster recovery

## Benefits
- Zero-downtime deployments
- Scalable architecture
- Automated deployment pipeline
- Infrastructure as code
- Security best practices
- Comprehensive monitoring
- Easy rollback capabilities
- Multi-environment support