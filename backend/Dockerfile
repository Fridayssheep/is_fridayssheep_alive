FROM golang:latest AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM debian:13-slim
WORKDIR /app

# 复制编译好的二进制文件
COPY --from=builder /app/main .

# 配置默认时区为亚洲/上海
RUN apt-get update && apt-get install -y tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 8080

CMD ["./main"]
