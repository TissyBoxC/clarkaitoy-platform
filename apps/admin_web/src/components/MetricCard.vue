<script setup lang="ts">
defineProps<{
  label: string
  value: string
  detail: string
}>()
</script>

<template>
  <article class="metric-card">
    <p class="metric-label">{{ label }}</p>
    <p class="metric-value">{{ value }}</p>
    <p class="metric-detail">{{ detail }}</p>
  </article>
</template>

<style scoped>
.metric-card {
  padding: 18px;
  border: 1px solid #d8e2dd;
  border-radius: 8px;
  background: #ffffff;
}

.metric-label,
.metric-detail {
  margin: 0;
  color: #61736c;
  font-size: 13px;
}

.metric-value {
  margin: 8px 0;
  color: #183c34;
  font-size: 28px;
  font-weight: 700;
}
</style>
