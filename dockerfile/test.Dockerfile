# 构建阶段
FROM golang:1.23 AS builder

# 设置工作目录
WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./
RUN go mod download

# 复制整个项目代码
COPY . .

# 构建user-service服务
RUN CGO_ENABLED=0 GOOS=linux go build -o test ./cmd/test

# 运行阶段
FROM alpine:3.18

# 安装基本工具
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 从构建阶段复制编译好的应用
COPY --from=builder /app/test .

# 暴露应用端口
EXPOSE 8081

# 运行应用
CMD ["./test"]
