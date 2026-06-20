FROM golang:1.25.7-alpine AS builder
LABEL org.opencontainers.image.title="Budget-Planner"\
      org.opencontainers.image.description="service for budget planning and financial management"\
      org.opencontainers.image.version="1.0.0"\
      org.opencontainers.image.authors="dmitriysyworov1986.com@gmail.com"\
      org.opencontainers.image.source="https://github.com/DmitriySyworov"
WORKDIR /app
COPY shared/go.mod shared/go.sum ./shared/
COPY services/budget-planner-service/go.mod services/budget-planner-service/go.sum ./services/budget-planner-service/
RUN go work init ./shared ./services/budget-planner-service
RUN go mod download -C services/budget-planner-service
COPY services/budget-planner-service/ ./services/budget-planner-service/
COPY shared/ ./shared/
WORKDIR /app/services/budget-planner-service
RUN go build -o /app/services/budget-planner-service/budget-planner-app ./cmd/app/main.go
RUN go build -o /app/services/budget-planner-service/migrator ./cmd/migration/auto.go
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/services/budget-planner-service/budget-planner-app .
COPY --from=builder /app/services/budget-planner-service/migrator .
ENTRYPOINT ["./budget-planner-app"]