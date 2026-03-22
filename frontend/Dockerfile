# 构建阶段
FROM node:22-bookworm-slim AS builder

WORKDIR /app
COPY package*.json ./
RUN npm install

COPY . .

# 接收 docker-compose 传进来的构建参数并设为环境变量供 Vite 编译使用
ARG VITE_REFRESH_INTERVAL
ENV VITE_REFRESH_INTERVAL=$VITE_REFRESH_INTERVAL

# 如果是通过 Nginx 转发，用相对路径（空字符串）让它默认访问根域名的 /api 的路由
ENV VITE_BACKEND_URL=""
RUN npm run build

# 运行阶段
FROM nginx:latest

# 移除默认的 nginx 配置
RUN rm /etc/nginx/conf.d/default.conf

# 拷贝自定义的 nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 拷贝构建好的前端文件到 Nginx
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
