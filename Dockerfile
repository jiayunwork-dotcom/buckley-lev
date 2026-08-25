# buckley-lev 项目构建镜像
# 单阶段：官方 Go 镜像自带完整工具链，编译缓存留在镜像内。
FROM golang:1.21-alpine
ENV GOTOOLCHAIN=local
ENV CGO_ENABLED=0

WORKDIR /app

# 先复制依赖文件并下载（利用 Docker 缓存；纯标准库，下载为空操作）。
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/server .
EXPOSE 8080
CMD ["/app/bin/server", "-http", ":8080"]
