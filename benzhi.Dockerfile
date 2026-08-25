# 官方 Go 镜像，自带完整工具链
FROM golang:1.21-alpine
ENV GOTOOLCHAIN=local
ENV CGO_ENABLED=0

WORKDIR /app

# 先复制依赖文件并下载依赖（利用 Docker 缓存，也保证容器内离线可用）
COPY go.mod go.sum ./
RUN go mod download

# 复制所有项目文件
COPY . .

RUN go build -o /app/bin/server .
EXPOSE 8080
CMD ["/app/bin/server", "-http", ":8080"]
