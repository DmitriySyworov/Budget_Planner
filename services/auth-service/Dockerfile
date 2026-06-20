FROM golang:1.25.7-alpine AS builder
LABEL org.opencontainers.image.title="Auth-User-Service"\
      org.opencontainers.image.description="service for authorizing users and performing actions on user accounts"\
      org.opencontainers.image.version="1.0.0"\
      org.opencontainers.image.authors="dmitriysyworov1986.com@gmail.com"\
      org.opencontainers.image.source="https://github.com/DmitriySyworov"
WORKDIR /app
COPY shared/go.mod shared/go.sum ./shared/
COPY services/auth-service/go.mod services/auth-service/go.sum ./services/auth-service/
RUN go work init ./shared ./services/auth-service
RUN go mod download -C services/auth-service
COPY services/auth-service/ ./services/auth-service/
COPY shared/ ./shared/
WORKDIR /app/services/auth-service
RUN go build -o /app/services/auth-service/auth-user-app ./cmd/app/main.go
RUN go build -o /app/services/auth-service/migrator ./cmd/migration/auto.go
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/services/auth-service/auth-user-app .
COPY --from=builder /app/services/auth-service/migrator .
ENTRYPOINT ["./auth-user-app"]