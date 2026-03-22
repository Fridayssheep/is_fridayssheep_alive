# Is Fridayssheep Alive？

一个用于视奸个人工作状态的专属监控仪表盘。实时展示给你的队友你当前是在爆肝写代码、网上冲浪看构式小视频，还是在听歌摸鱼。

- 由于这个B不懂前端，所以前端部分的代码90%都是Vibe code完成，如果遇到不合理的错误还请指出

## 核心特性

- **本地活动追踪**：集成 [ActivityWatch](https://activitywatch.net/)，实时展示当前Windows上正在使用的软件与窗口标题。
- **音乐播放状态**：集成 Last.fm，动态展示当前正在播放的歌曲及专辑封面
- **硬件状态监控**：通过 SSH 协议连接工作站，实时拉取 CPU、内存、GPU 以及 Ollama 正在运行的大模型状态。
- **GitHub 状态**：展示最新的 GitHub 活跃情况。
- **个性化**：
  - 基于当前时间的动态问候语（早/中/晚/深夜）。
  - 根据进程名称（如 Code / Steam / 浏览器）自定义状态标签（正在爆肝 / 享受游戏 / 网上冲浪）及其主题色。
  - 配置通过外置 `config.json` 热加载，**修改无需重新编译或重启**。

## 技术栈

- **前端**：Vue 3 + Vite + Element Plus (`Node.js 22`)
- **后端**：Go 1.26+ (标准库原生 HTTP 服务 + goroutine 并发高频缓存架构)
- **部署**：Docker + Docker Compose (基于 `debian:13-slim` 与 `nginx:latest`)

---

## 快速开始与部署

本项目已完全容器化，使用 Docker Compose 可以一键完成双端构建与部署。
### 1. 准备环境
你的Windows需要安装并运行 [ActivityWatch](https://activitywatch.net/) 并配置**允许局域网**访问api接口以供后端抓取活动数据。同时确保目标工作站机器（可以是同一台或局域网内的另一台**Linux**）开启了 SSH 服务以供后端抓取硬件状态。

完成后，将该项目克隆到你的机器上：

```bash
git clone https://github.com/Fridayssheep/is_fridayssheep_alive.git
cd is_fridayssheep_alive
```
### 2. 准备环境变量与配置文件
项目的根目录下需要提供一个 `.env` 文件。你可以参照根目录下的`.env.example`并填写你真实的配置：

```env
# 目标机器 SSH 连接信息 (用于抓取硬件数据)
SSH_HOST=192.168.1.100
SSH_PORT=22
SSH_USER=root
SSH_PASSWORD=your_password
# SSH_KEY_PATH=/root/.ssh/id_rsa # 如果使用密钥登录

# 个人账号及 API 地址
GITHUB_USERNAME=Fridayssheep
OLLAMA_API_URL=http://192.168.1.100:11434
# 指向你 Windows 上 ActivityWatch API 的地址
ACTIVITYWATCH_URL=http://192.168.1.10:5600
```

将src内部的 `config.json` 文件拷贝到根目录下，或者直接在 `docker-compose.yml` 中将宿主机的配置文件映射进容器以便后续热更新：

### 2. 配置你的docker-compose.yml
```yaml
services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: frisheep-backend
    restart: unless-stopped
    env_file:
      - .env
    ports:
      # 如果你希望后端直接对宿主机暴露可以取消这行注释
      # - "8080:8080"
      - "8080" # 只在 Docker 网络内暴露
    # 将包含 SSH 私钥的目录映射以便程序可以使用它
    # volumes:
    #   - ~/.ssh/id_rsa:/root/.ssh/id_rsa:ro

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
      args:
        - VITE_REFRESH_INTERVAL=${VITE_REFRESH_INTERVAL:-5}
        ## 修改env文件中的VITE_REFRESH_INTERVAL值或是直接修改${VITE_REFRESH_INTERVAL:-5}为数字可以调整前端自动重拉远端设备状态的间隔时间，单位为秒
    container_name: frisheep-frontend
    restart: unless-stopped
    ports:
      # 对外暴露前端网页的端口，例如这里映射为本机的 5601 端口
      - "5601:80"
    volumes:
      # 将宿主机的配置文件映射进容器，后续即可不用重启镜像就直接热生效修改
      - ./config.json:/usr/share/nginx/html/config.json:ro
      # 如果你想自定义背景图，也可以将背景图放在宿主机上并映射进容器，例如：
      #- ./backgrounds.webp:/usr/share/nginx/html/backgrounds.webp:ro
    depends_on:
      - backend

```
### 3. 一键启动
在项目根目录运行：

```bash
docker-compose up -d --build
```
等待镜像构建并启动完成后，在浏览器访问 `http://localhost:5601` 即可

---

## 配置文件说明 (`config.json`)

系统通过解析前端的 `config.json` 来决定 UI 展现形式：

```json
{
  "activityRules": [
    {
      "match": ["Code", "idea", "Cursor"], 
      "label": "正在爆肝",
      "color": "#67C23A"
    },
    {
      "match": ["Steam", "Genshin Impact"],
      "label": "打游戏",
      "color": "#F56C6C"
    }
  ],
  "activityDefault": {
    "label": "正常使用",
    "color": "#909399"
  },
  "timeGreetingRules": [
    { "start": "00:00", "end": "05:59", "text": "夜深了，注意休息哦🌙", "color": "#909399" }
  ]
}
```
*提示：`match` 字段支持模糊匹配（无视大小写），只要进程名字中包含了数组里的关键词即会自动应用配置的主题色。*

## License
本项目遵循MIT License开源协议，欢迎任何形式的贡献和使用！如果你有任何建议或想法，欢迎提交Issue或Pull Request。
