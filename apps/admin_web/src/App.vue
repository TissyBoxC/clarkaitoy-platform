<script setup lang="ts">
import AdminLayout from '@/layouts/AdminLayout.vue'
import { useRoute } from 'vue-router'
import { computed } from 'vue'

const route = useRoute()
const isLoginRoute = computed(() => route.meta.public === true)
</script>

<template>
  <Transition name="page" mode="out-in">
    <RouterView v-if="isLoginRoute" />
    <AdminLayout v-else />
  </Transition>
</template>
