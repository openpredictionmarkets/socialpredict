FROM golang:1.25.1-alpine3.22 AS builder

WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/server .

FROM alpine:3.22.1

RUN addgroup -S app && adduser -S -G app app
COPY --from=builder /app/server /usr/local/bin/server
RUN chown app:app /usr/local/bin/server
USER app
ENV BACKEND_PORT=8080
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/server"]
