FROM golang:1.26.4-alpine3.22 AS builder

WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/server .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/mcpserver ./cmd/mcpserver

FROM alpine:3.24.1

RUN addgroup -S app && adduser -S -G app app \
	&& mkdir -p /data/socialpredict/uploads/social-share \
	&& chown -R app:app /data/socialpredict
COPY --from=builder /app/server /usr/local/bin/server
COPY --from=builder /app/mcpserver /usr/local/bin/mcpserver
RUN chown app:app /usr/local/bin/server /usr/local/bin/mcpserver
USER app
ENV BACKEND_PORT=8080
ENV MCP_PORT=8081
EXPOSE 8080 8081
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
	CMD wget -q -O - "http://127.0.0.1:${BACKEND_PORT:-8080}/health" | grep -qx live
ENTRYPOINT ["/usr/local/bin/server"]
