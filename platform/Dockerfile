FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /server /server
# 创建 ConfigMap 挂载点
USER root
RUN mkdir -p /app/configs && chown nonroot:nonroot /app/configs
USER nonroot:nonroot
ENTRYPOINT ["/server"]
