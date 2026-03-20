<template>
  <div class="panel">
    <el-card shadow="hover" class="glow-card delay-1">
      <template #header>
        <div class="card-header">
          <div style="display: flex; align-items: center; gap: 8px;">
            <el-icon><Discount /></el-icon> 🐙 GitHub 提交状态
          </div>
          <el-tag :type="status?.github?.has_recent_activity ? 'success' : 'info'" effect="dark" round>
            {{ status?.github?.has_recent_activity ? '卷王在线' : '近期在摸鱼' }}
          </el-tag>
        </div>
      </template>

      <div v-if="loading" class="loading">读取中...</div>
      <div v-else-if="status?.github" class="hw-list">
        <div class="hw-item">
          <el-icon><Notification /></el-icon>
          <div class="hw-content">
            <div class="hw-label">最近一次活动类型</div>
            <div class="hw-value">
              <el-tag size="small" effect="plain">{{ status.github.last_activity_type }}</el-tag>
            </div>
          </div>
        </div>

        <div class="hw-item">
          <el-icon><FolderOpened /></el-icon>
          <div class="hw-content">
            <div class="hw-label">操作的代码库</div>
            <div class="hw-value">
              <el-link 
                type="primary" 
                :href="'https://github.com/' + status.github.last_activity_repo" 
                target="_blank"
                style="font-weight: bold; font-size: 1rem;"
              >
                {{ status.github.last_activity_repo }}
              </el-link>
            </div>
          </div>
        </div>

        <div class="hw-item">
          <el-icon><Clock /></el-icon>
          <div class="hw-content">
            <div class="hw-label">最后活动时间</div>
            <div class="hw-value">{{ new Date(status.github.last_activity_time).toLocaleString() }}</div>
          </div>
        </div>
      </div>
      
      <el-empty v-else description="暂时没有获取到 GitHub 数据..." :image-size="60">
        <template #image>
          <el-icon size="60" color="#c0c4cc"><Discount /></el-icon>
        </template>
      </el-empty>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { Discount, Notification, FolderOpened, Clock } from '@element-plus/icons-vue'

const status = ref(null)
const loading = ref(true)

const fetchStatus = async () => {
  try {
    const res = await axios.get('/api/status')
    status.value = res.data
  } catch (error) {
    console.error('Failed to fetch github status:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchStatus()
})
</script>

<style scoped>
.panel {
  padding: 10px 0;
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
  opacity: 0;
  animation: fadeInUp 0.7s cubic-bezier(0.2, 0.8, 0.2, 1) forwards;
}
.delay-1 { animation-delay: 0.1s; }

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 1.1rem;
}

/* --------------- 列表信息 --------------- */
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
  color: var(--el-color-primary, #4abeb6);
  background: rgba(74, 190, 182, 0.15);
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

:deep(.el-empty) {
  padding: 20px 0;
}
:deep(.el-empty__description) {
  margin-top: 10px;
  font-size: 0.9rem;
  color: #999;
}
</style>
