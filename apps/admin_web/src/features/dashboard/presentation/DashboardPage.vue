<script setup lang="ts">
import MetricCard from '@/components/MetricCard.vue'
</script>

<template>
  <section>
    <header class="page-header">
      <div>
        <p class="eyebrow">运营概览</p>
        <h1>平台状态</h1>
        <p class="page-description">集中查看家庭、设备和 AI 服务的关键运行状态。</p>
      </div>
    </header>

    <div class="metric-grid">
      <MetricCard label="家庭" value="--" detail="等待家庭服务同步" />
      <MetricCard label="在线设备" value="--" detail="等待设备状态同步" />
      <MetricCard label="今日 AI 对话" value="--" detail="等待网关统计同步" />
      <MetricCard label="待发布更新" value="--" detail="等待发布记录同步" />
    </div>
  </section>
</template>

<style scoped>
.page-header {
  margin-bottom: 24px;
}

.eyebrow {
  margin: 0 0 6px;
  color: var(--sprout-pink-strong);
  font-size: 13px;
  font-weight: 700;
}

h1 {
  margin: 0;
  color: var(--sprout-text);
  font-size: 28px;
}

.page-description {
  max-width: 680px;
  margin: 8px 0 0;
  color: var(--sprout-text-muted);
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

.metric-grid > :nth-child(2) {
  animation-delay: 55ms;
}

.metric-grid > :nth-child(3) {
  animation-delay: 110ms;
}

.metric-grid > :nth-child(4) {
  animation-delay: 165ms;
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
