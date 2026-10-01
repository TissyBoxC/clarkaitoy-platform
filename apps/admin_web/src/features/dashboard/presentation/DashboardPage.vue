<script setup lang="ts">
import MetricCard from '@/components/MetricCard.vue'
</script>

<template>
  <section>
    <header class="page-header">
      <div>
        <p class="eyebrow">运营概览</p>
        <h1>平台状态</h1>
      </div>
    </header>

    <div class="metric-grid">
      <MetricCard label="家庭" value="--" detail="等待业务服务接入" />
      <MetricCard label="在线设备" value="--" detail="等待设备状态接入" />
      <MetricCard label="今日 AI 对话" value="--" detail="等待网关统计接入" />
      <MetricCard label="待发布更新" value="--" detail="等待发布记录接入" />
    </div>
  </section>
</template>

<style scoped>
.page-header {
  margin-bottom: 24px;
}

.eyebrow {
  margin: 0 0 6px;
  color: #61736c;
  font-size: 13px;
  font-weight: 600;
  text-transform: uppercase;
}

h1 {
  margin: 0;
  color: #183c34;
  font-size: 28px;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

@media (max-width: 1080px) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 560px) {
  .metric-grid {
    grid-template-columns: 1fr;
  }
}
</style>
