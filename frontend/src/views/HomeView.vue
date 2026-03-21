<template>
  <div class="home-container">
    <el-card class="hero-card" shadow="never">
      <div class="hero-content">
        <h1 class="title">Fridayssheep还活着吗</h1>
        <p class="subtitle">看看这BYD是不是又去摸鱼了……</p>
        <div class="intro">
          <p>欢迎来到我的个人服务监控面板！</p>
          <p>可以视奸我的工作站运行状态和 GitHub 摸鱼情况。</p>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { Microphone } from '@element-plus/icons-vue'
</script>

<style scoped>
.home-container {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-top: 40px;
}

.hero-card {
  width: 100%;
  max-width: 800px;
  text-align: center;
  padding: 40px 20px;
  background: rgba(255, 255, 255, 0.75) !important;
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

.desc {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: rgba(74, 190, 182, 0.15);
  color: var(--el-color-primary);
  padding: 10px 20px;
  border-radius: 20px;
  font-weight: bold;
  margin-top: 10px;
}

.intro {
  margin: 20px 0;
  color: #555;
  line-height: 1.8;
  font-size: 1.1rem;
}
</style>
