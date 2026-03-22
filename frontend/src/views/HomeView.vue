<template>
  <div class="home-container">
    <!-- 顶部介绍原大块 -->
    <el-card class="hero-card" shadow="never">
      <div class="hero-content">
        <h1 class="title">Fridayssheep还活着吗</h1>
        <p class="subtitle">看看这BYD是不是又去摸鱼了……</p>
        <div class="intro">
          <p>欢迎来到我的个人服务监控面板！</p>
          <p>可以监（shi）督（jian）我正在干什么……</p>
        </div>
        <div class="current-time-container">
          <div class="current-time">当前时间：{{ currentTime }}</div>
          <div v-if="timeGreeting" class="time-greeting" :style="{ color: timeGreeting.color }">
            {{ timeGreeting.text }}
          </div>
        </div>
      </div>
    </el-card>

    <div class="status-grid">
      <!-- 动态数据块 -->
      <el-card class="status-block" shadow="hover">
        <template #header>
          <div class="block-header">
            <div class="block-header-title">
              <el-icon><Monitor /></el-icon>
              <span>Fridayssheep的主力机</span>
            </div>
            <!-- 用户自定义标签 -->
            <div v-if="activityData && activityData.app" class="status-tag" :style="{ color: activityStatus.color, backgroundColor: activityStatus.color + '1A' }">
              {{ activityStatus.label }}
            </div>
          </div>
        </template>
        <div class="activity-section" v-if="activityData">
          <div class="activity-card-inner" v-if="activityData && activityData.app">
            <div class="activity-info">
              <el-icon class="activity-icon" size="28" :color="activityData.is_active ? '#67C23A' : '#909399'"><Coffee v-if="!activityData.is_active" /><Monitor v-else /></el-icon>
              <div class="activity-text">
                <div class="app-name">{{ activityData.app }}</div>
                <div class="window-title">{{ activityData.title }}</div>
              </div>
            </div>
          </div>
          <el-alert v-else type="info" center show-icon :closable="false" title="暂无活动数据，也许去睡觉了" />
        </div>
        <el-skeleton v-else :rows="2" animated />
      </el-card>

      <!-- 音乐播放状态块 -->
      <el-card class="status-block music-block" shadow="hover">
        <template #header>
          <div class="block-header">
            <el-icon><Headset /></el-icon>
            <span>正在听什么</span>
          </div>
        </template>
        <div id="lfm-widget" class="lfm-container" :class="{ 'is-playing': lastfmData.isPlaying }">
          <div class="lfm-body">
              <div class="lfm-cover-wrapper">
                  <img v-if="lastfmData.cover" :src="lastfmData.cover" id="lfm-cover" alt="Cover">
                  <div v-else id="lfm-placeholder" class="lfm-img-placeholder">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <circle cx="12" cy="12" r="10"></circle>
                          <circle cx="12" cy="12" r="3"></circle>
                      </svg>
                  </div>
                  <div class="lfm-status-dot"></div>
              </div>
              <div class="lfm-info">
                  <div id="lfm-status-text" class="lfm-label" :style="{ color: lastfmData.isPlaying ? '#39c5bb' : 'var(--text-color-meta)' }">
                    {{ lastfmData.statusText || '正在同步...' }}
                  </div>
                  <div id="lfm-track" class="lfm-title">{{ lastfmData.track || '-' }}</div>
                  <div id="lfm-artist" class="lfm-subtitle">{{ lastfmData.artist || '-' }}</div>
              </div>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, reactive, computed } from 'vue'
import { Microphone, Monitor, Coffee, Headset } from '@element-plus/icons-vue'
import axios from 'axios'
import dayjs from 'dayjs'
import isBetween from 'dayjs/plugin/isBetween'
dayjs.extend(isBetween)

const activityData = ref(null)
const appConfig = ref(null)
const currentTime = ref('')
const currentDayjs = ref(dayjs())
let timer = null
let clockTimer = null

// Last.fm 数据状态
const lastfmData = reactive({
  track: '',
  artist: '',
  cover: '',
  isPlaying: false,
  statusText: ''
})

const lfmConfig = {
  apiKey: 'e7a1a1304a85c2cf4201f88e65d3fa8f',
  user: 'Fridayssheep',
  metingApi: 'https://apimusic.frp.fridayssheep.top/api',
  interval: 30000 
}

function timeAgo(uts) {
    const seconds = Math.floor(Date.now() / 1000) - parseInt(uts);
    if (seconds < 60) return '刚刚播放';
    const mins = Math.floor(seconds / 60);
    return mins < 60 ? `${mins} 分钟前` : `${Math.floor(mins / 60)} 小时前`;
}

const activityStatus = computed(() => {
  if (!activityData.value || !appConfig.value) {
    return { label: '...', color: '#909399' }
  }

  // 1. 如果不活跃，优先显示摸鱼
  if (!activityData.value.is_active) {
    return { label: '似乎离线了', color: '#909399' }
  }

  // 2. 如果活跃，根据 config.json 进行匹配进程名称
  const appName = activityData.value.app || ''
  const rules = appConfig.value.activityRules || []
  
  for (const rule of rules) {
    // 支持模糊匹配
    if (rule.match.some(m => appName.toLowerCase().includes(m.toLowerCase()))) {
      return { label: rule.label, color: rule.color || '#67C23A' }
    }
  }

  // 3. Fallback 到默认配置
  return appConfig.value.activityDefault || { label: '正常活动', color: '#409EFF' }
})

const timeGreeting = computed(() => {
  if (!appConfig.value || !appConfig.value.timeGreetingRules) return null
  
  const rules = appConfig.value.timeGreetingRules
  const now = currentDayjs.value
  
  for (const rule of rules) {
    if (!rule.start || !rule.end) continue
    
    // 把 hh:mm 格式提取并附加上当前的日期转换为能够比较的 dayjs 对象
    const startDate = dayjs(`${now.format('YYYY-MM-DD')} ${rule.start}:00`)
    const endDate = dayjs(`${now.format('YYYY-MM-DD')} ${rule.end}:59`)
    
    if (now.isBetween(startDate, endDate, null, '[]')) {
      return { text: rule.text, color: rule.color }
    }
  }
  return { text: '你好呀', color: '#666' }
})

const fetchConfig = async () => {
  try {
    // 作为打包后的静态资源，直接请求根目录的 config.json
    const res = await axios.get('/config.json')
    appConfig.value = res.data
  } catch (error) {
    console.error("加载配置文件失败", error)
  }
}

const fetchActivity = async () => {
  try {
    const res = await axios.get(`/api/activity`)
    activityData.value = res.data
  } catch (error) {
    console.error("加载活动数据失败", error)
  }
}

const fetchLastfm = async () => {
  try {
    const res = await axios.get(`https://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks&user=${lfmConfig.user}&api_key=${lfmConfig.apiKey}&format=json&limit=1`)
    if (!res.data.recenttracks || !res.data.recenttracks.track[0]) return
    
    const track = res.data.recenttracks.track[0]
    lastfmData.track = track.name
    lastfmData.artist = track.artist['#text']
    lastfmData.isPlaying = track['@attr'] && track['@attr'].nowplaying === 'true'
    
    if (lastfmData.isPlaying) {
      lastfmData.statusText = '正在播放'
    } else {
      lastfmData.statusText = track.date ? timeAgo(track.date.uts) : '最近播放'
    }

    // 获取封面
    try {
      const metingRes = await axios.get(`${lfmConfig.metingApi}?server=netease&type=search&id=${encodeURIComponent(track.name + ' ' + track.artist['#text'])}`)
      if (Array.isArray(metingRes.data) && metingRes.data.length > 0) {
        lastfmData.cover = metingRes.data[0].pic || metingRes.data[0].cover
      } else {
        lastfmData.cover = ''
      }
    } catch (e) {
      lastfmData.cover = ''
    }
  } catch (e) {
    console.error("Last.fm 抓取失败", e)
  }
}

onMounted(() => {
  // 钟表
  clockTimer = setInterval(() => {
    currentDayjs.value = dayjs()
    currentTime.value = currentDayjs.value.format('YYYY年MM月DD日 HH:mm:ss')
  }, 1000)
  currentDayjs.value = dayjs()
  currentTime.value = currentDayjs.value.format('YYYY年MM月DD日 HH:mm:ss')

  fetchConfig()
  fetchActivity()
  fetchLastfm()
  
  const refreshInterval = parseInt(import.meta.env.VITE_REFRESH_INTERVAL || '5') * 1000
  timer = setInterval(() => {
    fetchActivity()
  }, refreshInterval)
  
  // Last.fm 单独的定时器
  setInterval(fetchLastfm, lfmConfig.interval)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
  if (clockTimer) clearInterval(clockTimer)
})
</script>

<style scoped>
.home-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-top: 40px;
  gap: 30px;
  width: 100%;
  overflow-x: hidden;
}

.hero-card {
  width: 100%;
  text-align: center;
  padding: 30px 20px;
  background: rgba(255, 255, 255, 0.75) !important;
  border-radius: 12px;
}

.hero-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 15px;
}

.title {
  font-size: 2.8rem;
  color: #333;
  margin: 0;
  font-weight: 800;
  letter-spacing: 2px;
}

.subtitle {
  font-size: 1.2rem;
  color: #666;
  margin: 0;
}

.intro {
  margin: 20px 0 10px 0;
  color: #555;
  line-height: 1.8;
  font-size: 1.1rem;
}

.current-time-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  margin-top: 10px;
}

.current-time {
  font-size: 1rem;
  color: #888;
  font-family: monospace;
  background: rgba(0, 0, 0, 0.05);
  padding: 6px 16px;
  border-radius: 20px;
}

.time-greeting {
  font-size: 0.95rem;
  font-weight: 600;
  letter-spacing: 1px;
}

/* 动态栅格布局 */
.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
  gap: 20px;
  width: 100%;
}

.status-block {
  background: rgba(255, 255, 255, 0.8) !important;
  border-radius: 12px;
}

:deep(.music-block .el-card__body) {
  padding: 0;
  overflow: hidden;
}

.block-header {
  display: flex;
  align-items: center;
  justify-content: space-between; /* 使得标题和标签分布在两侧 */
  width: 100%;
}

.block-header-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: bold;
  font-size: 1.1rem;
  color: #333;
}

/* Activity 样式调整 */
.activity-section {
  width: 100%;
}
.activity-card-inner {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.activity-info {
  display: flex;
  align-items: center;
  gap: 15px;
}
.activity-text {
  display: flex;
  flex-direction: column;
}
.app-name {
  font-weight: bold;
  font-size: 1.1rem;
  color: #409EFF;
}
.window-title {
  font-size: 0.9rem;
  color: #666;
  word-break: break-word; /* 允许长标题换行 */
  white-space: normal;
  margin-top: 4px;
  line-height: 1.4;
}
.status-tag {
  font-size: 0.85rem; /* 字号缩小一点适应标题栏 */
  padding: 4px 10px;
  border-radius: 20px;
  font-weight: bold;
}

/* Last.fm 组件样式迁移 */
.lfm-container { 
    position: relative;
  width: 100%;
  box-sizing: border-box;
  padding: 20px;
    background: transparent; 
    border-radius: 0 0 12px 12px; /* 贴合卡片底部圆角 */
    overflow: hidden;
    z-index: 1;
}

.lfm-container::after {
    content: "";
    position: absolute;
    top: 0; 
    left: -100%; 
    width: 100%; 
    height: 100%;
    background: linear-gradient(
        90deg, 
        transparent 0%, 
        rgba(57, 197, 187, 0.01) 30%, 
        rgba(57, 197, 187, 0.2) 50%, 
        rgba(57, 197, 187, 0.01) 70%, 
        transparent 100%
    );
    mix-blend-mode: screen; 
    pointer-events: none;
    display: none; 
}

.lfm-container.is-playing::after {
    display: block;
    animation: dsScanPrecision 5s linear infinite;
}

@keyframes dsScanPrecision {
    0% { transform: translateX(0); }
    90% { transform: translateX(200%); } 
    100% { transform: translateX(200%); } 
}

.lfm-body { display: flex; align-items: center; gap: 14px; position: relative; z-index: 2; }
.lfm-cover-wrapper { position: relative; flex-shrink: 0; width: 60px; height: 60px; }

#lfm-cover, .lfm-img-placeholder { 
    width: 100%; height: 100%; border-radius: 8px; 
    object-fit: cover; border: 1px solid rgba(57, 197, 187, 0.15);
}

.lfm-img-placeholder { 
    display: flex; align-items: center; justify-content: center;
    background: rgba(57, 197, 187, 0.05); color: #39c5bb;
}

.lfm-status-dot { 
    position: absolute; bottom: -2px; right: -2px; width: 12px; height: 12px; 
    background: #39c5bb; border: 2px solid #fff; 
    border-radius: 50%; display: none; box-shadow: 0 0 6px #39c5bb;
}
.is-playing .lfm-status-dot { display: block; animation: dsPulse 3s infinite; }

@keyframes dsPulse { 
    0% { box-shadow: 0 0 0 0 rgba(57, 197, 187, 0.5); } 
    70% { box-shadow: 0 0 0 6px rgba(57, 197, 187, 0); } 
    100% { box-shadow: 0 0 0 0 rgba(57, 197, 187, 0); } 
}

.lfm-info { min-width: 0; flex: 1; text-align: left;}
.lfm-label { font-size: 11px; text-transform: uppercase; margin-bottom: 4px; font-weight: bold; letter-spacing: 1px; }
.lfm-title { font-weight: 600; font-size: 16px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: #333; }
.lfm-subtitle { font-size: 13px; color: #666; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; margin-top: 2px;}
</style>
