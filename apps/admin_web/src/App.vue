<script setup lang="ts">
import AdminLayout from '@/layouts/AdminLayout.vue'
import { useRoute } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()
const isLoginRoute = computed(() => route.meta.public === true)
</script>

<template>
  <RouterView v-if="isLoginRoute" />
  <AdminLayout v-else />
</template>

<style>
* {
  box-sizing: border-box;
}

body {
  margin: 0;
  color: #4a2e3b;
  background: #fffbfc;
  font-family: Inter, 'Microsoft YaHei', sans-serif;
}
</style>
