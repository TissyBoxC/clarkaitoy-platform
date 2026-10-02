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
  position: relative;
  overflow: hidden;
  padding: 20px;
  border: 1px solid var(--sprout-outline);
  border-radius: var(--sprout-radius-card);
  background: #ffffff;
  box-shadow: 0 8px 24px rgb(194 91 128 / 5%);
  transition:
    transform var(--sprout-duration-slow) var(--sprout-ease-out),
    border-color var(--sprout-duration-base) ease,
    box-shadow var(--sprout-duration-slow) ease;
}

.metric-card::before {
  position: absolute;
  top: 0;
  left: 22px;
  width: 42px;
  height: 4px;
  border-radius: 0 0 999px 999px;
  background: linear-gradient(90deg, #f7a8bf, #ffe066);
  content: '';
}

.metric-card:hover {
  border-color: #e9a5b8;
  box-shadow: var(--sprout-shadow-hover);
  transform: translateY(-3px);
}

.metric-label,
.metric-detail {
  margin: 0;
  color: var(--sprout-text-muted);
  font-size: 13px;
}

.metric-value {
  margin: 8px 0;
  color: var(--sprout-text);
  font-size: 28px;
  font-weight: 700;
}
</style>
