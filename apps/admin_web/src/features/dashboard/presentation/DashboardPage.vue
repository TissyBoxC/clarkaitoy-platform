<script setup lang="ts">
import MetricCard from '@/components/MetricCard.vue'
</script>

<template>
  <section>
    <header class="page-header">
      <div>
        <p class="eyebrow">Operations overview</p>
        <h1>Platform status</h1>
      </div>
    </header>

    <div class="metric-grid">
      <MetricCard label="Families" value="--" detail="Awaiting platform API" />
      <MetricCard label="Devices online" value="--" detail="Awaiting telemetry API" />
      <MetricCard label="AI calls today" value="--" detail="Awaiting gateway metrics" />
      <MetricCard label="OTA pending" value="--" detail="Awaiting release API" />
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
