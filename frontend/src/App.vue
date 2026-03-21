<template>
  <div class="app-layout">
    <el-container>
      <el-header class="glass-header">
        <div class="header-content">
          <div class="logo">Fridayssheep似了吗</div>
          
          <div class="nav-container">
            <el-tabs v-model="activeRoute" @tab-click="handleTabClick" class="nav-tabs">
              <el-tab-pane name="/">
                <template #label>
                  <span class="tab-label">
                    <el-icon><HomeFilled /></el-icon> 首页
                  </span>
                </template>
              </el-tab-pane>
              <el-tab-pane name="/github">
                <template #label>
                  <span class="tab-label">
                    <el-icon><Discount /></el-icon> GitHub 状态
                  </span>
                </template>
              </el-tab-pane>
              <el-tab-pane name="/workstation">
                <template #label>
                  <span class="tab-label">
                    <el-icon><Platform /></el-icon> 工作站状态
                  </span>
                </template>
              </el-tab-pane>
            </el-tabs>
          </div>

        </div>
      </el-header>
      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { HomeFilled, Discount, Platform } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()

const activeRoute = ref(route.path)

watch(() => route.path, (newPath) => {
  activeRoute.value = newPath
})

const handleTabClick = (tab) => {
  if (tab.paneName && tab.paneName !== route.path) {
    router.push(tab.paneName)
  }
}
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  margin: 0;
  padding: 0;
}

.glass-header {
  padding: 0;
  background: rgba(255, 255, 255, 0.45);
  backdrop-filter: blur(15px);
  -webkit-backdrop-filter: blur(15px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  position: sticky;
  top: 0;
  z-index: 1000;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

.logo {
  font-size: 1.2rem;
  font-weight: bold;
  color: #333;
  letter-spacing: 1px;
}

.nav-container {
  display: flex;
  align-items: flex-end;
  height: 60px;
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 1.05rem;
}

/* 去掉底部多余背景线，只保留滑动的 indicator */
:deep(.el-tabs__nav-wrap::after) {
  display: none; 
}

:deep(.el-tabs__header) {
  margin: 0 !important;
}

/* 控制 item 高度与对齐 */
:deep(.el-tabs__item) {
  height: 60px;
  line-height: 60px;
  padding: 0 15px !important;
}

/* 滑动横线加粗加圆角 */
:deep(.el-tabs__active-bar) {
  height: 3px;
  border-radius: 3px;
  background-color: var(--el-color-primary);
  bottom: 0px;
}

/* 标签状态切换过渡 */
:deep(.el-tabs__item) {
  transition: color 0.3s cubic-bezier(0.645, 0.045, 0.355, 1), font-weight 0.3s;
}
:deep(.el-tabs__item:hover), :deep(.el-tabs__item.is-active) {
  color: var(--el-color-primary) !important;
  font-weight: bold;
}

.main-content {
  padding: 40px 20px;
  max-width: 1000px;
  margin: 0 auto;
  width: 100%;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.4s ease, transform 0.4s ease;
}

.fade-enter-from {
  opacity: 0;
  transform: translateY(10px);
}
.fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
