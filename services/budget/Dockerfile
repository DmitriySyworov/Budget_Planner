FROM golang:1.25.7-alpine AS builder
LABEL org.opencontainers.image.title="Budget-Planner"\
      org.opencontainers.image.description="service for budget planning and financial management"\
      org.opencontainers.image.version="1.0.0"\
      org.opencontainers.image.authors="dmitriysyworov1986.com@gmail.com"\
      org.opencontainers.image.source="https://github.com/DmitriySyworov"
WORKDIR /app
COPY shared/go.mod shared/go.sum ./shared/
COPY services/budget/go.mod services/budget/go.sum ./services/budget/
RUN go work init ./shared ./services/budget
RUN go mod download -C services/budget
COPY services/budget/ ./services/budget/
COPY shared/ ./shared/
WORKDIR /app/services/budget
RUN go build -o /app/services/budget/budget-app ./cmd/app/main.go
RUN go build -o /app/services/budget/migrator ./cmd/migration/auto.go
FROM alpine:3.23.5
WORKDIR /app
COPY --from=builder /app/services/budget/budget-app .
COPY --from=builder /app/services/budget/migrator .
ENTRYPOINT ["./budget-app"]