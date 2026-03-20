<template>
  <div class="panel">
    <!-- 全局状态栏 -->
    <el-card shadow="hover" class="status-card glow-card" style="margin-bottom: 20px; animation-delay: 0s;">
      <div class="status-header">
        <div class="title">
          <el-icon><Platform /></el-icon>
          工作站核心节点
        </div>
        <div class="status-right">
          <div class="countdown-badge">
            <el-icon :class="{ 'is-loading': statusLoading }"><RefreshRight /></el-icon>
            <span>{{ countdown }}s 后刷新</span>
          </div>
          <el-tag :type="status?.status === 'ok' ? 'success' : 'danger'" effect="dark" round>
            {{ status?.status === 'ok' ? '节点在线' : '数据异常或离线' }}
          </el-tag>
        </div>
      </div>
    </el-card>

    <el-row :gutter="20">
      <!-- 硬件静态信息 -->
      <el-col :xs="24" :md="12" style="margin-bottom: 20px;">
        <el-card shadow="hover" class="glow-card h-100 delay-1">
          <template #header>
            <div class="card-header"><el-icon><Opportunity /></el-icon> 配置信息</div>
          </template>
          <div v-if="hwLoading" class="loading">读取中...</div>
          <div v-else class="hw-list">
            <div class="hw-item">
              <el-icon><Monitor /></el-icon>
              <div class="hw-content">
                <div class="hw-label">操作系统</div>
                <div class="hw-value">{{ hardwareInfo?.os || '未知 OS' }}</div>
              </div>
            </div>
            <div class="hw-item">
              <el-icon><Cpu /></el-icon>
              <div class="hw-content">
                <div class="hw-label">处理器 (CPU)</div>
                <div class="hw-value">{{ hardwareInfo?.cpu || '未知 CPU' }}</div>
              </div>
            </div>
            <div class="hw-item">
              <el-icon><Coin /></el-icon>
              <div class="hw-content">
                <div class="hw-label">物理内存</div>
                <div class="hw-value">{{ hardwareInfo?.mem_total || '未知' }}</div>
              </div>
            </div>
            <div class="hw-item">
              <el-icon><VideoCamera /></el-icon>
              <div class="hw-content">
                <div class="hw-label">图形处理器 (GPU)</div>
                <div class="hw-value" v-if="hardwareInfo?.gpus && hardwareInfo.gpus.length">
                  <div v-for="(gpu, idx) in hardwareInfo.gpus" :key="idx" style="margin-bottom: 4px; font-size: 0.9rem; line-height: 1.4;">
                    <el-tag size="small" type="info" style="margin-right: 6px;">GPU {{ idx }}</el-tag>
                    {{ gpu }}
                  </div>
                </div>
                <div class="hw-value" v-else>未检测到独立显卡</div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 实时性能监控 (动态波浪填充动画) -->
      <el-col :xs="24" :md="12" style="margin-bottom: 20px;">
        <el-card shadow="hover" class="glow-card h-100 delay-2">
          <template #header>
            <div class="card-header"><el-icon><Odometer /></el-icon> 实时占用看板</div>
          </template>
          <div v-if="!status && statusLoading" class="loading">连接中...</div>
          <div v-else class="kuma-metrics">

            <!-- CPU -->
            <div class="kuma-service">
              <div class="kuma-service-left">
                <el-tag :type="status?.system ? getTagType(status.system.cpu_percent) : 'info'" effect="dark" round class="percentage-tag">
                  {{ status?.system ? status.system.cpu_percent + '%' : '离线' }}
                </el-tag>
                <div class="kuma-service-name">
                  <el-icon><Cpu /></el-icon> CPU 占用
                </div>
              </div>
              <div class="kuma-bars-container">
                <div class="kuma-bars">
                  <el-tooltip
                    v-for="i in totalBars" :key="'cpu-'+i"
                    :content="'当前总体占用: ' + (status?.system ? status.system.cpu_percent : 0) + '%'"
                    placement="top"
                    :show-after="50"
                  >
                    <div class="kuma-bar-bg">
                      <div class="kuma-bar-fill" :style="{
                        height: getBarHeight(displayCpu, i),
                        backgroundColor: getBarColor(displayCpu),
                        transitionDelay: cpuDirection === 'up' ? `${i * 0.015}s` : `${(totalBars - i) * 0.015}s`
                      }"></div>
                    </div>
                  </el-tooltip>
                </div>
                <div class="kuma-bars-label">
                  <span>0%</span>
                  <span>100%</span>
                </div>
              </div>
            </div>

            <!-- Memory -->
            <div class="kuma-service" style="border-bottom: none; margin-bottom: 0;">
              <div class="kuma-service-left">
                <el-tag :type="status?.system ? getTagType(status.system.mem_percent) : 'info'" effect="dark" round class="percentage-tag">
                  {{ status?.system ? status.system.mem_percent.toFixed(1) + '%' : '离线' }}
                </el-tag>
                <div class="kuma-service-name">
                  <el-icon><Coin /></el-icon> RAM 占用
                </div>
              </div>
              <div class="kuma-bars-container">
                <div class="kuma-bars">
                  <el-tooltip
                    v-for="i in totalBars" :key="'mem-'+i"
                    :content="'当前总体使用率: ' + (status?.system ? status.system.mem_percent.toFixed(1) : 0) + '%'"
                    placement="top"
                    :show-after="50"
                  >
                    <div class="kuma-bar-bg">
                      <div class="kuma-bar-fill" :style="{
                        height: getBarHeight(displayMem, i),
                        backgroundColor: getBarColor(displayMem),
                        transitionDelay: memDirection === 'up' ? `${i * 0.015}s` : `${(totalBars - i) * 0.015}s`
                      }"></div>
                    </div>
                  </el-tooltip>
                </div>
                <div class="kuma-bars-label">
                  <span v-if="status?.system" style="color: #666; font-weight: bold;">
                    已用 {{ formatBytes(status.system.mem_used) }} / {{ formatBytes(status.system.mem_total) }}
                  </span>
                  <span v-else>0%</span>
                  <span>100%</span>
                </div>
              </div>
            </div>

          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <!-- GPU -->
      <el-col :xs="24" :md="12" style="margin-bottom: 20px;">
        <el-card shadow="hover" class="glow-card h-100 delay-3">
          <template #header>
            <div class="card-header"><el-icon><VideoCamera /></el-icon> GPU 运行状态</div>
          </template>
          
          <div v-if="status?.gpus && status.gpus.length" class="kuma-metrics" style="padding-top: 0; gap: 15px;">
            <div v-for="(gpu, idx) in status.gpus" :key="'gpu-cfg-'+idx" style="border-bottom: 1px dashed rgba(0,0,0,0.08); padding-bottom: 15px; margin-bottom: 5px;">
              <div style="font-size: 0.95rem; font-weight: bold; margin-bottom: 12px; color: #555; display: flex; align-items: center;">
                <el-icon style="margin-right: 4px;"><VideoCamera /></el-icon> GPU 核心 {{ idx }}
                <el-tag size="small" type="danger" round style="margin-left:10px;">
                  <div style="display: flex; align-items: center; gap: 4px;"><el-icon><Odometer /></el-icon> <span>{{ gpu.temperature }}</span></div>
                </el-tag>
                <el-tag size="small" type="warning" round style="margin-left:6px;">
                  <div style="display: flex; align-items: center; gap: 4px;"><el-icon><Lightning /></el-icon> <span>{{ gpu.power_draw }}</span></div>
                </el-tag>
              </div>

              <!-- GPU 利用率 -->
              <div class="kuma-service" style="border-bottom: none; margin-bottom: 10px; padding-bottom: 0;">
                <div class="kuma-service-left">
                  <el-tag :type="getTagType(gpu.utilization)" effect="dark" round class="percentage-tag">
                    {{ gpu.utilization }}
                  </el-tag>
                  <div class="kuma-service-name" style="font-size: 0.9rem;">计算</div>
                </div>
                <div class="kuma-bars-container">
                  <div class="kuma-bars" style="height: 25px;">
                    <el-tooltip
                      v-for="i in totalBars" :key="'gutil-'+idx+'-'+i"
                      :content="'核心占用: ' + gpu.utilization"
                      placement="top"
                      :show-after="50"
                    >
                      <div class="kuma-bar-bg">
                        <div class="kuma-bar-fill" :style="{
                          height: getBarHeight(displayGpus[idx]?.util, i),
                          backgroundColor: getBarColor(displayGpus[idx]?.util),
                          transitionDelay: displayGpus[idx]?.utilDir === 'up' ? `${i * 0.015}s` : `${(totalBars - i) * 0.015}s`
                        }"></div>
                      </div>
                    </el-tooltip>
                  </div>
                </div>
              </div>

              <!-- GPU 显存 -->
              <div class="kuma-service" style="border-bottom: none; margin-bottom: 0; padding-bottom: 0;">
                <div class="kuma-service-left">
                  <el-tag :type="getTagType((parseFloat(gpu.memory_used)/parseFloat(gpu.memory_total))*100)" effect="dark" round class="percentage-tag">
                    {{ ((parseFloat(gpu.memory_used)/parseFloat(gpu.memory_total))*100).toFixed(1) }}%
                  </el-tag>
                  <div class="kuma-service-name" style="font-size: 0.9rem;">显存</div>
                </div>
                <div class="kuma-bars-container">
                  <div class="kuma-bars" style="height: 25px;">
                    <el-tooltip
                      v-for="i in totalBars" :key="'gmem-'+idx+'-'+i"
                      :content="'VRAM使用: ' + gpu.memory_used + ' / ' + gpu.memory_total"
                      placement="top"
                      :show-after="50"
                    >
                      <div class="kuma-bar-bg">
                        <div class="kuma-bar-fill" :style="{
                          height: getBarHeight(displayGpus[idx]?.mem, i),
                          backgroundColor: getBarColor(displayGpus[idx]?.mem),
                          transitionDelay: displayGpus[idx]?.memDir === 'up' ? `${i * 0.015}s` : `${(totalBars - i) * 0.015}s`
                        }"></div>
                      </div>
                    </el-tooltip>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <el-empty v-else description="设备离线 或 尚未采集到 GPU 信息" :image-size="60">
            <template #image>
              <el-icon size="60" color="#c0c4cc"><VideoCamera /></el-icon>
            </template>
          </el-empty>
        </el-card>
      </el-col>

      <!-- Ollama -->
      <el-col :xs="24" :md="12" style="margin-bottom: 20px;">
        <el-card shadow="hover" class="glow-card h-100 delay-4">
          <template #header>
            <div class="card-header"><el-icon><MagicStick /></el-icon> 🦙 Ollama 服务</div>
          </template>
          
          <div v-if="status?.ollama && status.ollama.length > 0" class="hw-list" style="margin-top: 10px;">
            <div class="hw-item" v-for="(model, idx) in status.ollama" :key="'ollama-'+idx" style="padding-bottom: 15px; border-bottom: 1px solid #f0f0f0;">
              <el-icon style="font-size: 32px; color: #8e44ad; background: rgba(142, 68, 173, 0.1);"><Box /></el-icon>
              <div class="hw-content" style="flex: 1;">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px;">
                  <span class="hw-value" style="font-weight: bold; color: var(--el-color-primary);">{{ model.name }}</span>
                  <el-tag size="small" type="success" effect="dark" round>占用 {{ formatBytes(model.size_vram) }}</el-tag>
                </div>
                <div class="hw-label" style="display: flex; align-items: center; gap: 4px;">
                  <el-icon><Clock /></el-icon> 驻留释放: {{ new Date(model.expires_at).toLocaleString() }}
                </div>
              </div>
            </div>
          </div>
          <el-empty v-else description="当前无存活的模型或服务未连接" :image-size="60">
            <template #image>
              <el-icon size="60" color="#c0c4cc"><MagicStick /></el-icon>
            </template>
          </el-empty>
        </el-card>
      </el-col>
    </el-row>

  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import axios from 'axios'
import { Platform, Monitor, Cpu, Coin, VideoCamera, Odometer, Opportunity, MagicStick, RefreshRight, Box, Clock, Lightning } from '@element-plus/icons-vue'

const hardwareInfo = ref(null)
const hwLoading = ref(true)

const status = ref(null)
const statusLoading = ref(true)
const timer = ref(null)

// 刷新间隔与倒计时配置
const refreshInterval = parseInt(import.meta.env.VITE_REFRESH_INTERVAL || '5')
const countdown = ref(refreshInterval)
let countdownTimer = null

// 用于动画渲染的显式百分比和增减方向
const displayCpu = ref(0)
const cpuDirection = ref('up') // 'up' 表示涨或初始加载，'down' 表示降
const displayMem = ref(0)
const memDirection = ref('up')
const displayGpus = ref([]) // 用于记录各个 GPU 的独立进度动画 { util, utilDir, mem, memDir }

const totalBars = 40 // 总共 40 个刻度，每一个代表 2.5%

// 获取标签颜色的逻辑 (根据占用率决定: 绿、橙、红)
const getTagType = (val) => {
  if (val === null || val === undefined) return 'info'
  const num = parseFloat(val)
  if (isNaN(num)) return 'info'
  if (num < 60) return 'success'
  if (num < 85) return 'warning'
  return 'danger'
}

// 计算每一个单根柱子的填充高度 (支持平滑填充一半的格子)
const getBarHeight = (percent, i) => {
  if (percent === null || percent === undefined) return '0%'
  const step = 100 / totalBars
  const minForBar = (i - 1) * step
  const maxForBar = i * step
  
  if (percent >= maxForBar) return '100%' // 完全超过这个格子，直接填满
  if (percent <= minForBar) return '0%'  // 根本没到这个格子，0%
  // 处于当前格子中，计算部分填充的高度
  return ((percent - minForBar) / step * 100) + '%'
}

// 获取颜色的逻辑（根据整体百分点决定亮起方块的颜色）
const getBarColor = (val) => {
  if (val === null || val === undefined) return '#e2e8f0'
  if (val === 0) return '#e2e8f0'
  if (val < 60) return '#4abeb6' // 主题青绿
  if (val < 85) return '#fbbf24' // 橙黄警告
  return '#f56c6c'               // 红色危险
}

const formatBytes = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const fetchHardware = async () => {
  try {
    const res = await axios.get('/api/hardware')
    hardwareInfo.value = res.data
  } catch (error) {
    console.error('Failed to fetch hardware:', error)
  } finally {
    hwLoading.value = false
  }
}

let isFirstLoad = true

const fetchStatus = async () => {
  statusLoading.value = true
  try {
    const res = await axios.get('/api/status')
    status.value = res.data
    
    const applyData = () => {
      // 1. 处理 CPU & RAM
      if (res.data?.system) {
        const newCpu = res.data.system.cpu_percent
        const newMem = res.data.system.mem_percent
        cpuDirection.value = newCpu >= displayCpu.value ? 'up' : 'down'
        displayCpu.value = newCpu
        memDirection.value = newMem >= displayMem.value ? 'up' : 'down'
        displayMem.value = newMem
      } else {
        cpuDirection.value = 'down'
        displayCpu.value = 0
        memDirection.value = 'down'
        displayMem.value = 0
      }

      // 2. 处理多 GPU 的数据与分别的波浪动画
      if (res.data?.gpus && res.data.gpus.length) {
        if (displayGpus.value.length !== res.data.gpus.length) {
          displayGpus.value = res.data.gpus.map(() => ({ util: 0, utilDir: 'up', mem: 0, memDir: 'up' }))
        }
        res.data.gpus.forEach((gpu, idx) => {
          const newUtil = parseFloat(gpu.utilization) || 0
          const memUsed = parseFloat(gpu.memory_used) || 0
          const memTotal = parseFloat(gpu.memory_total) || 1
          const newMem = (memUsed / memTotal) * 100

          displayGpus.value[idx].utilDir = newUtil >= displayGpus.value[idx].util ? 'up' : 'down'
          displayGpus.value[idx].util = newUtil

          displayGpus.value[idx].memDir = newMem >= displayGpus.value[idx].mem ? 'up' : 'down'
          displayGpus.value[idx].mem = newMem
        })
      } else {
        displayGpus.value = []
      }
    }

    if (isFirstLoad) {
      isFirstLoad = false
      // 延迟600ms，等卡片的 fadeInUp 进场动画基本完成后再开始波浪填充
      setTimeout(applyData, 600)
    } else {
      applyData()
    }

  } catch (error) {
    console.error('Failed to fetch status:', error)
    // 失败当离线处理
    cpuDirection.value = 'down'
    displayCpu.value = 0
    memDirection.value = 'down'
    displayMem.value = 0
    displayGpus.value = []
  } finally {
    statusLoading.value = false
    countdown.value = refreshInterval // 请求完成后重置倒计时
  }
}

onMounted(() => {
  fetchHardware()
  fetchStatus()
  
  // 数据拉取定时器
  timer.value = setInterval(() => {
    fetchStatus()
  }, refreshInterval * 1000)

  // 倒计时的本地定时器
  countdownTimer = setInterval(() => {
    if (countdown.value > 0) {
      countdown.value--
    } else {
      countdown.value = refreshInterval
    }
  }, 1000)
})

onUnmounted(() => {
  if (timer.value) {
    clearInterval(timer.value)
  }
  if (countdownTimer) {
    clearInterval(countdownTimer)
  }
})
</script>

<style scoped>
.panel {
  padding: 10px 0;
}

.h-100 {
  height: 100%;
}

/* --------------- 进场悬浮动画 --------------- */
@keyframes fadeInUp {
  0% {
    opacity: 0;
    transform: translateY(30px);
  }
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}

.glow-card {
  opacity: 0; /* 初始隐藏 */
  animation: fadeInUp 0.7s cubic-bezier(0.2, 0.8, 0.2, 1) forwards;
}

.delay-1 { animation-delay: 0.1s; }
.delay-2 { animation-delay: 0.25s; }
.delay-3 { animation-delay: 0.4s; }
.delay-4 { animation-delay: 0.55s; }


/* --------------- 顶部状态与倒计时 --------------- */
.status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.status-header .title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 1.2rem;
  font-weight: bold;
  color: var(--el-color-primary);
}

.status-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.countdown-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.9rem;
  color: var(--el-color-primary);
  background: rgba(74, 190, 182, 0.15);
  padding: 4px 12px;
  border-radius: 20px;
  font-weight: bold;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.1rem;
}

/* --------------- 硬件信息 --------------- */
.hw-list {
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.hw-item {
  display: flex;
  align-items: center;
  gap: 15px;
}
.hw-item .el-icon {
  font-size: 24px;
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
  padding: 10px;
  border-radius: 12px;
}
.hw-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.hw-label {
  font-size: 0.85rem;
  color: #888;
}
.hw-value {
  font-size: 1rem;
  font-weight: 500;
  color: #333;
}

/* --------------- 进度条面板 --------------- */
.kuma-metrics {
  padding: 10px 0;
  display: flex;
  flex-direction: column;
  gap: 25px;
}

.kuma-service {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 25px;
  padding-bottom: 20px;
  border-bottom: 1px dashed rgba(0,0,0,0.08);
}

.kuma-service-left {
  display: flex;
  align-items: center;
  gap: 15px;
}
.percentage-tag {
  min-width: 60px;
  text-align: center;
  font-weight: bold;
}
.kuma-service-name {
  font-weight: bold;
  color: #555;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 1.05rem;
}

.kuma-bars-container {
  width: 100%;
  display: flex;
  flex-direction: column;
}

.kuma-bars {
  display: flex;
  gap: 4px;
  width: 100%;
  height: 35px;
}

.kuma-bar-bg {
  flex: 1;
  background-color: #e2e8f0;
  border-radius: 4px;
  min-width: 4px;
  overflow: hidden;
  display: flex;
  align-items: flex-end;
  cursor: pointer;
}

.kuma-bar-fill {
  width: 100%;
  border-radius: 4px; 
  /* 放慢生长速度并增加弹性, 延迟在行内样式中动态结算 */
  transition: height 0.6s cubic-bezier(0.34, 1.56, 0.64, 1), background-color 0.6s ease;
}

.kuma-bar-bg:hover .kuma-bar-fill {
  filter: brightness(0.85);
}

.kuma-bars-label {
  width: 100%;
  display: flex;
  justify-content: space-between;
  margin-top: 8px;
  font-size: 0.8rem;
  color: #a0aec0;
}

.json-box {
  background: rgba(0,0,0,0.03);
  padding: 15px;
  border-radius: 8px;
  overflow-x: auto;
  font-size: 0.85rem;
}

:deep(.el-empty) {
  padding: 20px 0;
}
:deep(.el-empty__description) {
  margin-top: 10px;
  font-size: 0.9rem;
  color: #999;
}
</style>
