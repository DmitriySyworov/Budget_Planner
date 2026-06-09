FROM golang:1.25.7-alpine AS builder
LABEL org.opencontainers.image.title="Budget-Planner"\
      org.opencontainers.image.description="service for budget planning and financial management"\
      org.opencontainers.image.version="1.0.0"\
      org.opencontainers.image.authors="dmitriysyworov1986.com@gmail.com"\
      org.opencontainers.image.source="https://github.com/DmitriySyworov"
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/budget-planner-app ./cmd/app/main.go
RUN go build -o /app/migrator ./cmd/migration/auto.go
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/budget-planner-app .
COPY --from=builder /app/migrator .
ENTRYPOINT ["./budget-planner-app"]